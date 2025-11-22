import { useState, useCallback } from 'react';
import { FileUploadResponse, uploadFileToS3 } from '../utils/uploadFileToS3';

interface UseFileUploadReturn {
  selectedFile: File | null;
  previewUrl: string | null;
  isUploading: boolean;
  handleFileSelect: (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleRemoveFile: () => void;
  // 🚨 upload 함수 시그니처 변경 (workspaceId 추가)
  upload: (
    workspaceId: string,
    category: 'project' | 'board' | 'comment' | 'chat',
  ) => Promise<FileUploadResponse | null>;
  setInitialFile: (fileUrl: string | null, fileName: string | null) => void;
}

export const useFileUpload = (): UseFileUploadReturn => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  // ... handleFileSelect, handleRemoveFile, setInitialFile 로직은 기존과 동일 ...
  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      if (file.type.startsWith('image/')) {
        setPreviewUrl(URL.createObjectURL(file));
      } else {
        setPreviewUrl(null);
      }
    }
  }, []);

  const handleRemoveFile = useCallback(() => {
    setSelectedFile(null);
    setPreviewUrl(null);
  }, []);

  const setInitialFile = useCallback((fileUrl: string | null, fileName: string | null) => {
    if (fileUrl) setPreviewUrl(fileUrl); // fileUrl이 곧 미리보기 URL이거나 다운로드 URL이라 가정
  }, []);

  // 💡 수정된 upload 함수
  const upload = async (
    workspaceId: string,
    category: 'project' | 'board' | 'comment' | 'chat',
  ) => {
    if (!selectedFile) return null;

    setIsUploading(true);
    try {
      // 서비스 함수 호출 시 workspaceId 전달
      const data = await uploadFileToS3(selectedFile, workspaceId, category);
      return data;
    } catch (error) {
      console.error('Upload hook error:', error);
      throw error;
    } finally {
      setIsUploading(false);
    }
  };

  return {
    selectedFile,
    previewUrl,
    isUploading,
    handleFileSelect,
    handleRemoveFile,
    upload,
    setInitialFile,
  };
};
