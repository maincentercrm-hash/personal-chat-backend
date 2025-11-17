-- ตรวจสอบ message ที่เป็น cursor
SELECT id, created_at, content, message_type 
FROM messages 
WHERE id = 'c53720dc-cfea-4fc9-a707-cb1a74fbea10';

-- ตรวจสอบ messages ที่อยู่รอบๆ cursor (10 messages ก่อนและหลัง)
SELECT id, created_at, content, message_type 
FROM messages 
WHERE conversation_id = '69cd966b-c0f4-44bf-ae6f-f08eaf501e20'
ORDER BY created_at DESC, id DESC 
LIMIT 30;
