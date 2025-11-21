// src/components/modals/board/ProjectManageModal.tsx

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { X, Calendar, Paperclip, Download, Edit2, BarChart3, Lock, Loader2 } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import { createProject, updateProject, getBoardsByProject } from '../../../api/board/boardService'; // 💡 [추가] getBoardsByProject import
import { ProjectResponse, BoardResponse } from '../../../types/board'; // 💡 [추가] BoardResponse import
import { formatDate } from '../../../utils/date';
import { IROLES } from '../../../types/common';
import Portal from '../../common/Portal';

/**
 * 모달 모드 타입 정의
 * 'create': 프로젝트 생성 폼
 * 'detail': 상세 보기 (읽기 전용)
 * 'edit': 상세 보기 중 수정 폼
 */
type ProjectModalMode = 'create' | 'detail' | 'edit';

interface ProjectManageModalProps {
  workspaceId: string;
  project?: ProjectResponse;
  onClose: () => void;
  onProjectSaved: () => void;
  onProjectCreated?: (createObj: ProjectResponse) => void;
  // 💡 [추가] 현재 사용자의 역할 (권한 제어용)
  userRole: IROLES;
  // 💡 [추가] 초기 모드 설정 (헤더의 상세 보기 버튼에서 시작할 때 사용)
  initialMode: ProjectModalMode;
}

// 💡 [추가] 파일 다운로드 핸들러 (재사용)
const handleFileDownload = (fileUrl: string, fileName: string) => {
  if (!fileUrl) return;

  const link = document.createElement('a');
  link.href = fileUrl;
  link.setAttribute('download', fileName);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

export const ProjectManageModal: React.FC<ProjectManageModalProps> = ({
  workspaceId,
  project,
  onClose,
  onProjectSaved,
  onProjectCreated,
  userRole,
  initialMode = 'create',
}) => {
  const { theme } = useTheme();
  const isExistingProject = !!project;

  // 💡 [통합] 현재 모달의 모드 상태
  const [mode, setMode] = useState<ProjectModalMode>(isExistingProject ? initialMode : 'create');

  // Form state
  const [name, setName] = useState(project?.name || '');
  const [description, setDescription] = useState(project?.description || '');
  // 💡 [추가] 마감일 (dueDate) 상태: ISO 문자열에서 YYYY-MM-DD 형식으로 변환하여 저장
  const [dueDate, setDueDate] = useState(project?.dueDate ? project.dueDate.substring(0, 10) : '');

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 [추가] 프로젝트 보드 상태 및 로딩
  const [boards, setBoards] = useState<BoardResponse[]>([]);
  const [isBoardsLoading, setIsBoardsLoading] = useState(false);

  // 💡 [권한 체크] OWNER 또는 ADMIN/ORGANIZER만 수정 권한을 가집니다.
  const canEdit = useMemo(() => {
    return (
      isExistingProject &&
      (userRole === 'OWNER' || userRole === 'ADMIN' || userRole === 'ORGANIZER')
    );
  }, [isExistingProject, userRole]);

  // project prop이 변경되거나 mode가 detail로 돌아가면 폼 리셋
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description || '');
      setDueDate(project.dueDate ? project.dueDate.substring(0, 10) : '');
    } else if (mode === 'create') {
      setName('');
      setDescription('');
      setDueDate('');
    }
    setError(null);
  }, [project, mode]);

  // 💡 [추가] 프로젝트 보드 API 호출 로직
  const fetchBoards = useCallback(async () => {
    if (!project || mode !== 'detail') {
      setBoards([]);
      return;
    }
    setIsBoardsLoading(true);
    try {
      // API 호출: http://localhost:8000/api/boards/project/{projectId}
      const response = await getBoardsByProject(project.projectId);
      // 💡 [가정] API 응답 구조: { data: BoardResponse[] }
      setBoards(response || []);
    } catch (err) {
      console.error('❌ Failed to fetch boards for statistics:', err);
      setBoards([]);
    } finally {
      setIsBoardsLoading(false);
    }
  }, [project, mode]);

  // mode가 'detail'로 변경될 때마다 보드 데이터 로드
  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  // 💡 [추가] 프로젝트 통계 계산 (실제 BoardResponse 타입에 맞게 수정 필요)
  const projectStats = useMemo(() => {
    const totalBoards = boards.length;

    // ⚠️ [임시 로직] BoardResponse 타입이 정의되지 않았으므로 임시로 status 필드를 가정
    const inProgressBoards = boards.filter((b) => (b as any).status === 'IN_PROGRESS').length;
    const delayedBoards = boards.filter((b) => (b as any).isDelayed).length; // 지연된 보드를 판단하는 로직 필요

    return {
      totalBoards,
      inProgressBoards,
      delayedBoards,
    };
  }, [boards]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError('프로젝트 이름은 필수입니다.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      if (mode === 'edit' && project) {
        // --- ✏️ 편집 로직 ---
        await updateProject(project.projectId, {
          name: name.trim(),
          description: description.trim() || undefined,
          dueDate: dueDate || undefined, // 💡 [추가] 마감일 추가
        });
        alert(`✅ ${name} 프로젝트가 수정되었습니다!`);
        onProjectSaved();
        setMode('detail'); // 수정 후 상세 보기로 돌아가기
      } else if (mode === 'create') {
        // --- ✨ 생성 로직 ---
        const newProjectResponse: ProjectResponse = await createProject({
          workspaceId: workspaceId,
          name: name.trim(),
          description: description.trim() || undefined,
          dueDate: dueDate || undefined, // 💡 [추가] 생성 시 마감일 추가
        });

        alert(`✅ ${name} 프로젝트가 생성되었습니다!`);

        if (newProjectResponse) {
          onProjectCreated?.(newProjectResponse);
        }
        onProjectSaved();
        onClose(); // 생성 성공 후 모달 닫기
      } else {
        // Detail 모드에서 Submit 버튼이 잘못 눌린 경우
        return;
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error(
        mode === 'create' ? '❌ 프로젝트 생성 실패:' : '❌ 프로젝트 수정 실패:',
        errorMsg,
      );
      setError(
        errorMsg ||
          (mode === 'create' ? '프로젝트 생성에 실패했습니다.' : '프로젝트 수정에 실패했습니다.'),
      );
    } finally {
      setIsLoading(false);
    }
  };

  // 💡 [추가] 모달 제목 설정
  const modalTitle = useMemo(() => {
    switch (mode) {
      case 'create':
        return '새 프로젝트 만들기';
      case 'edit':
        return `${project?.name || '프로젝트'} 수정`;
      case 'detail':
      default:
        return `${project?.name || '프로젝트'} 상세 정보`;
    }
  }, [mode, project?.name]);

  // 💡 [추가] 파일 정보 가져오기 (가정)
  const fileUrl = (project as any)?.fileUrl;
  const fileName = (project as any)?.fileName || 'project_file_attachment';

  // ----------------------------------------------------
  // 🎨 Detail / Edit Mode 렌더링
  // ----------------------------------------------------
  const renderDetailOrEditContent = () => (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Name / Title */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">
          프로젝트 이름 <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={mode === 'detail' || isLoading}
          placeholder="예: Wealist 서비스 개발"
          className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm ${
            mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
          }`}
          maxLength={100}
          autoFocus
        />
      </div>

      {/* 💡 [추가] Due Date */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">프로젝트 마감일</label>
        <input
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          disabled={mode === 'detail' || isLoading}
          className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm ${
            mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
          }`}
        />
      </div>

      {/* Description */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">프로젝트 설명</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          disabled={mode === 'detail' || isLoading}
          placeholder="프로젝트에 대한 간단한 설명을 입력하세요"
          className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none ${
            mode === 'detail' ? 'bg-gray-100 text-gray-700' : 'bg-white'
          }`}
          rows={3}
          maxLength={500}
        />
      </div>

      {/* Files (Detail/Edit Mode에서만 표시) */}
      {mode !== 'create' && (
        <div className="pt-0">
          <label className="block text-sm font-semibold text-gray-700 mb-2 flex items-center gap-1">
            <Paperclip className="w-4 h-4 text-blue-500" />
            첨부 파일
          </label>
          <div className="p-2 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-between text-sm">
            <span className="text-gray-700 truncate flex items-center gap-1">
              {fileUrl ? (
                <span className="text-gray-700">{fileName}</span>
              ) : (
                <span className="text-gray-500">첨부 파일 없음</span>
              )}
            </span>

            {fileUrl ? (
              <button
                type="button"
                onClick={() => handleFileDownload(fileUrl, fileName)}
                className="flex items-center gap-1 text-blue-600 hover:text-blue-700 transition font-medium ml-2 flex-shrink-0"
                disabled={isLoading}
              >
                <Download className="w-4 h-4" />
                <span className="text-xs">다운로드</span>
              </button>
            ) : (
              <span className="text-gray-400 text-xs flex-shrink-0">첨부 가능</span>
            )}
          </div>
        </div>
      )}

      {/* 💡 [추가] 통계 섹션 (Detail Mode에서만 표시) */}
      {mode === 'detail' && (
        <>
          <h3 className="text-md font-bold text-gray-800 flex items-center gap-2 pt-2">
            <BarChart3 className="w-5 h-5 text-indigo-500" /> 프로젝트 현황
          </h3>

          {isBoardsLoading ? (
            <div className="flex justify-center items-center py-4 text-gray-500">
              <Loader2 className="w-5 h-5 animate-spin mr-2" />
              통계 데이터 로드 중...
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-3 text-center">
              <div className="p-3 bg-indigo-50 rounded-lg border border-indigo-200">
                <p className="text-3xl font-bold text-indigo-700">{projectStats.totalBoards}</p>
                <p className="text-xs text-indigo-500 mt-1">총 보드 수</p>
              </div>
              <div className="p-3 bg-green-50 rounded-lg border border-green-200">
                <p className="text-3xl font-bold text-green-700">{projectStats.inProgressBoards}</p>
                <p className="text-xs text-green-500 mt-1">진행 중 보드</p>
              </div>
              <div className="p-3 bg-red-50 rounded-lg border border-red-200">
                <p className="text-3xl font-bold text-red-700">{projectStats.delayedBoards}</p>
                <p className="text-xs text-red-500 mt-1">지연 보드</p>
              </div>
            </div>
          )}
        </>
      )}

      {/* Timestamps (Detail/Edit Mode일 때만 표시) */}
      {mode !== 'create' && project && (
        <div className="text-xs text-gray-500 space-y-1 pt-2 border-t border-gray-100">
          <div className="flex items-center gap-2">
            <Calendar className="w-3 h-3 text-gray-400" />
            <span className="font-semibold text-gray-700">생성일:</span>
            {formatDate(project.createdAt)}
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="w-3 h-3 text-gray-400" />
            <span className="font-semibold text-gray-700">수정일:</span>
            {formatDate(project.updatedAt)}
          </div>
          {/* 💡 [추가] 마감일 표시 */}
          {project?.dueDate && (
            <div className="flex items-center gap-2">
              <Calendar className="w-3 h-3 text-gray-400" />
              <span className="font-semibold text-gray-700">마감일:</span>
              {formatDate(project?.dueDate)}
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-3 pt-4">
        {/* 💡 [수정] 1. 저장/생성 버튼 (Primary Action - Left) */}
        {(mode === 'edit' || mode === 'create') && (
          <button
            type="submit"
            className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
              isLoading ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            disabled={isLoading}
          >
            {isLoading
              ? mode === 'edit'
                ? '저장 중...'
                : '생성 중...'
              : mode === 'edit'
              ? '수정 내용 저장'
              : '프로젝트 만들기'}
          </button>
        )}

        {/* 💡 [수정] 2. 취소 버튼 (Secondary Action - Right) */}
        {mode === 'edit' && (
          <button
            type="button"
            onClick={() => setMode('detail')}
            className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
            disabled={isLoading}
          >
            취소 (상세 보기로)
          </button>
        )}

        {/* 💡 [수정] 3. 닫기 버튼 (Detail Mode 전용) */}
        {mode === 'detail' && (
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 font-semibold rounded-lg hover:bg-gray-300 transition"
          >
            닫기
          </button>
        )}
      </div>
    </form>
  );

  // ----------------------------------------------------
  // 🎨 Create Mode 렌더링 (유지)
  // ----------------------------------------------------
  const renderCreateContent = () => (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Name */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">
          프로젝트 이름 <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="예: Wealist 서비스 개발"
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          disabled={isLoading}
          maxLength={100}
          autoFocus
        />
      </div>

      {/* 💡 [추가] Due Date in Create Mode */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">프로젝트 마감일</label>
        <input
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          disabled={isLoading}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
        />
      </div>

      {/* Description */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">프로젝트 설명</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="프로젝트에 대한 간단한 설명을 입력하세요"
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none"
          rows={3}
          disabled={isLoading}
          maxLength={500}
        />
      </div>

      {/* Actions (Create Mode) */}
      <div className="flex gap-3 pt-2">
        {/* 💡 [수정] 생성 버튼이 왼쪽에 오도록 순서 변경 */}
        <button
          type="submit"
          className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
            isLoading ? 'opacity-50 cursor-not-allowed' : ''
          }`}
          disabled={isLoading}
        >
          {isLoading ? '생성 중...' : '프로젝트 만들기'}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
          disabled={isLoading}
        >
          취소
        </button>
      </div>
    </form>
  );

  return (
    <Portal>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[100]"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-md ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-4 pb-2 border-b border-gray-200">
            <div className="flex">
              <h2 className="text-xl font-bold text-gray-800">{modalTitle}</h2>

              {/* 💡 Detail/Edit Mode 전환 버튼 */}
              {mode !== 'create' && canEdit && (
                <div className="flex items-center gap-3">
                  {mode === 'detail' ? (
                    <button
                      onClick={() => setMode('edit')}
                      title="프로젝트 수정"
                      className="p-2 rounded-full hover:bg-yellow-50 text-yellow-600 transition"
                    >
                      <Edit2 className="w-5 h-5" />
                    </button>
                  ) : (
                    <button
                      onClick={() => setMode('detail')}
                      title="수정 취소"
                      className="p-2 rounded-full hover:bg-gray-100 text-gray-600 transition"
                    >
                      <Lock className="w-5 h-5" />
                    </button>
                  )}
                </div>
              )}
            </div>
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Project Owner Info (생성 모드 제외) */}
          {mode !== 'create' && project && (
            <div className="mb-4 p-3 bg-gray-50 rounded-lg">
              <div className="text-xs text-gray-500 mb-1">프로젝트 담당자</div>
              <div className="text-sm font-medium text-gray-700">{project?.ownerName}</div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
              {error}
            </div>
          )}

          {/* Content Render */}
          {mode === 'create' ? renderCreateContent() : renderDetailOrEditContent()}
        </div>
      </div>
    </Portal>
  );
};
