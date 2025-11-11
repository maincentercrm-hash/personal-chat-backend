// application/serviceimpl/auth_service.go
package serviceimpl

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5" // เปลี่ยนเป็น v5
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo           repository.UserRepository
	refreshTokenRepo   repository.RefreshTokenRepository
	tokenBlacklistRepo repository.TokenBlacklistRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenBlacklistRepo repository.TokenBlacklistRepository,
) service.AuthService {
	return &authService{
		userRepo:           userRepo,
		refreshTokenRepo:   refreshTokenRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
	}
}

func (s *authService) Register(username, password, email, displayName string) (*models.User, string, string, error) {

	// ตรวจสอบข้อมูลขั้นต่ำ
	if username == "" || password == "" {
		return nil, "", "", errors.New("username and password are required")
	}

	// ตรวจสอบว่ามีชื่อผู้ใช้นี้ในระบบแล้วหรือไม่
	existingUser, err := s.userRepo.FindByUsername(username)
	if err == nil && existingUser != nil {
		return nil, "", "", errors.New("username already exists")
	}

	// เข้ารหัสรหัสผ่าน
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", errors.New("failed to hash password")
	}

	// สร้างข้อมูลผู้ใช้
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		DisplayName:  displayName,
		Status:       "active",
		CreatedAt:    time.Now(),
		Settings:     types.JSONB{},
	}

	// บันทึกผู้ใช้
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", errors.New("failed to create user: " + err.Error())
	}

	// สร้าง tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, username)
	if err != nil {
		return nil, "", "", errors.New("failed to generate tokens: " + err.Error())
	}

	// บันทึก refresh token
	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		// อาจบันทึกข้อผิดพลาดที่นี่ แต่ยังคงดำเนินการต่อไป
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(username, password string) (*models.User, string, string, error) {
	// ตรวจสอบข้อมูลขั้นต่ำ
	if username == "" || password == "" {
		return nil, "", "", errors.New("username and password are required")
	}

	// ค้นหาผู้ใช้
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, "", "", errors.New("invalid username or password")
	}

	// ตรวจสอบรหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid username or password")
	}

	// อัปเดตเวลาใช้งานล่าสุด
	now := time.Now()
	user.LastActiveAt = &now
	if err := s.userRepo.Update(user); err != nil {
		// บันทึกข้อผิดพลาด แต่ไม่กระทบการเข้าสู่ระบบ
		log.Printf("Failed to update last_active_at: %v", err)
	}

	// สร้าง tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, username)
	if err != nil {
		return nil, "", "", errors.New("failed to generate tokens: " + err.Error())
	}

	// ยกเลิก refresh tokens เดิม
	if err := s.refreshTokenRepo.RevokeByUserID(user.ID); err != nil {
		// อาจบันทึกข้อผิดพลาดที่นี่ แต่ยังคงดำเนินการต่อไป
	}

	// บันทึก refresh token ใหม่
	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		// อาจบันทึกข้อผิดพลาดที่นี่ แต่ยังคงดำเนินการต่อไป
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(refreshTokenStr string) (string, string, error) {
	// ตรวจสอบ refresh token ในฐานข้อมูล
	refreshTokenModel, err := s.refreshTokenRepo.FindByToken(refreshTokenStr)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// ตรวจสอบว่า token หมดอายุหรือไม่
	if refreshTokenModel.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token expired")
	}

	// ค้นหาผู้ใช้
	user, err := s.userRepo.FindByID(refreshTokenModel.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// สร้าง tokens ใหม่
	accessToken, newRefreshToken, err := s.generateTokens(user.ID, user.Username)
	if err != nil {
		return "", "", errors.New("failed to generate tokens: " + err.Error())
	}

	// ยกเลิก refresh token เดิม
	if err := s.refreshTokenRepo.RevokeByUserID(user.ID); err != nil {
		// อาจบันทึกข้อผิดพลาดที่นี่ แต่ยังคงดำเนินการต่อไป
	}

	// อัปเดตเวลาใช้งานล่าสุด
	now := time.Now()
	user.LastActiveAt = &now
	s.userRepo.Update(user)

	// บันทึก refresh token ใหม่
	newRefreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.refreshTokenRepo.Create(newRefreshTokenModel); err != nil {
		// อาจบันทึกข้อผิดพลาดที่นี่ แต่ยังคงดำเนินการต่อไป
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) Logout(userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeByUserID(userID)
}

func (s *authService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *authService) generateTokens(userID uuid.UUID, username string) (string, string, error) {
	now := time.Now()

	// สร้าง Access Token (อายุสั้น)
	accessTokenClaims := jwt.MapClaims{
		"id":       userID,
		"username": username,
		"type":     "access",
		"exp":      now.Add(time.Hour * 24).Unix(), // หมดอายุใน 24 ชั่วโมง
		"iat":      now.Unix(),                     // เวลาที่ออกโทเคน
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	// สร้าง Refresh Token (อายุยาว)
	refreshTokenClaims := jwt.MapClaims{
		"id":       userID,
		"username": username,
		"type":     "refresh",
		"exp":      now.Add(time.Hour * 24 * 30).Unix(), // หมดอายุใน 30 วัน
		"iat":      now.Unix(),                          // เวลาที่ออกโทเคน
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	// เซ็น tokens
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-for-development-only"
	}

	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// เพิ่มเมธอด BlacklistToken
func (s *authService) BlacklistToken(userID uuid.UUID, token string) error {
	// ตรวจสอบว่า token อยู่ใน blacklist แล้วหรือไม่
	isBlacklisted, err := s.tokenBlacklistRepo.IsTokenBlacklisted(token)
	if err != nil {
		return err
	}

	if isBlacklisted {
		// ถ้ามีอยู่แล้ว ถือว่าสำเร็จ
		return nil
	}

	// ถ้ายังไม่มี ให้เพิ่มเข้าไป
	blacklist := &models.TokenBlacklist{
		Token:     token,
		UserID:    userID,
		ExpiredAt: time.Now().Add(time.Hour * 72), // ให้หมดอายุตามเวลาของ token
		CreatedAt: time.Now(),
	}

	return s.tokenBlacklistRepo.Create(blacklist)
}
