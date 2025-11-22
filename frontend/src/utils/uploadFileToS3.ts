import axios from 'axios'; // S3 업로드용 (인터셉터 없는 순수 axios 필요)
import { boardServiceClient } from '../api/apiConfig';

export interface FileUploadResponse {
  fileKey: string; // DB에 저장할 키
  fileName: string;
  // 필요하다면 s3Url 등 추가 가능
}

/**
 * Presigned URL 방식을 이용한 S3 파일 업로드
 */
export const uploadFileToS3 = async (
  file: File,
  workspaceId: string, // 🚨 API 명세에 필수라고 되어 있음
  category: 'project' | 'board' | 'comment' | 'chat' = 'project',
): Promise<FileUploadResponse> => {
  try {
    // 1️⃣ [Backend] Presigned URL 요청 (JSON)
    // entityType은 대문자로 변환 (PROJECT, BOARD 등)
    const presignedRes = await boardServiceClient.post('/attachments/presigned-url', {
      contentType: file.type,
      entityType: category.toUpperCase(),
      fileName: file.name,
      fileSize: file.size,
      workspaceId: workspaceId,
    });

    // 명세서 Response 구조: { data: { expiresIn, fileKey, uploadUrl }, requestId }
    const { uploadUrl, fileKey } = presignedRes.data.data;

    // 2️⃣ [S3] 실제 파일 업로드 (PUT)
    // 🚨 주의: 여기선 boardServiceClient를 쓰면 안됩니다 (BaseURL, Auth 헤더 등이 꼬임)
    // 순수한 axios를 사용하며, Content-Type을 파일과 맞춰야 합니다.
    await axios.put(uploadUrl, file, {
      headers: {
        'Content-Type': file.type,
      },
    });

    console.log('✅ S3 Upload Success:', fileKey);

    // 3️⃣ 결과 반환
    return {
      fileKey: fileKey, // 나중에 다운로드하거나 참조할 때 쓸 Key
      fileName: file.name,
    };
  } catch (error) {
    console.error('❌ File Upload Failed:', error);
    throw error;
  }
};
