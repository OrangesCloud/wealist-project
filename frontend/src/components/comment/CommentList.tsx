import { useState } from 'react';
import { Pencil, Trash2, Check, X } from 'lucide-react'; // 아이콘 사용 (없으면 텍스트로 대체 가능)
import { CommentResponse } from '../../types/board';
import { deleteComment, updateComment } from '../../api/board/boardService';
import { WorkspaceMemberResponse } from '../../types/user';
import { useUserLookup } from '../../hooks/useUserLookup';

// =============================================================================
// [Sub Component] 개별 댓글 아이템 (스스로 수정/삭제 모드 관리)
// =============================================================================
interface CommentItemProps {
  comment: CommentResponse;
  nickname: string;
  profileUrl: string | null;
  currentUserId: string; // 현재 로그인한 내 ID (버튼 노출 여부 판단용)
  onRefresh: () => void; // 수정/삭제 후 목록 새로고침 요청
}

const CommentItem = ({
  comment,
  nickname,
  profileUrl,
  currentUserId,
  onRefresh,
}: CommentItemProps) => {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isLoading, setIsLoading] = useState(false);

  // 내가 쓴 댓글인지 확인
  const isMyComment = comment.userId === currentUserId;

  // 닉네임 기반 배경색 (AvatarStack 로직 재사용)
  const getUserColor = (name: string) => {
    const colors = [
      'bg-indigo-500',
      'bg-pink-500',
      'bg-green-500',
      'bg-purple-500',
      'bg-yellow-500',
    ];
    return colors[name.length % colors.length];
  };

  // 수정 저장 핸들러
  const handleUpdate = async () => {
    if (!editContent.trim()) return;
    setIsLoading(true);
    try {
      await updateComment(comment.commentId, {
        boardId: comment.boardId, // API 스펙에 필요하다면
        content: editContent,
      });
      setIsEditing(false);
      onRefresh(); // 부모에게 데이터 갱신 요청
    } catch (error) {
      alert('댓글 수정에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  // 삭제 핸들러
  const handleDelete = async () => {
    if (!window.confirm('정말 삭제하시겠습니까?')) return;
    setIsLoading(true);
    try {
      await deleteComment(comment.commentId);
      onRefresh(); // 부모에게 데이터 갱신 요청
    } catch (error) {
      alert('댓글 삭제에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="p-3 bg-gray-100 border border-gray-200 rounded-lg group">
      <div className="flex items-start gap-2">
        {/* 아바타 영역 */}
        <div className="w-6 h-6 rounded-full flex-shrink-0 overflow-hidden ring-1 ring-white bg-gray-200 mt-1">
          {profileUrl ? (
            <img src={profileUrl} alt={nickname} className="w-full h-full object-cover" />
          ) : (
            <div
              className={`w-full h-full flex items-center justify-center text-white text-xs font-bold ${getUserColor(
                nickname,
              )}`}
            >
              {nickname?.[0] || '?'}
            </div>
          )}
        </div>

        {/* 컨텐츠 영역 */}
        <div className="flex-1 min-w-0">
          {/* 헤더 (이름 + 날짜 + 버튼들) */}
          <div className="flex items-center justify-between mb-1">
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold">{nickname}</span>
              <span className="text-[10px] text-gray-500">
                {/* formatDate 함수는 외부에 있다고 가정 */}
                {new Date(comment.createdAt).toLocaleDateString()}
              </span>
            </div>

            {/* 수정/삭제 버튼 (내 글이고, 수정 모드가 아닐 때만 보임) */}
            {isMyComment && !isEditing && (
              <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={() => setIsEditing(true)}
                  className="p-1 text-gray-400 hover:text-blue-500 hover:bg-gray-200 rounded"
                  title="수정"
                >
                  <Pencil size={12} />
                </button>
                <button
                  onClick={handleDelete}
                  className="p-1 text-gray-400 hover:text-red-500 hover:bg-gray-200 rounded"
                  title="삭제"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            )}
          </div>

          {/* 본문 (일반 모드 vs 수정 모드) */}
          {isEditing ? (
            <div className="mt-1">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className="w-full text-sm p-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500 resize-none"
                rows={2}
                disabled={isLoading}
              />
              <div className="flex justify-end gap-2 mt-2">
                <button
                  onClick={() => {
                    setIsEditing(false);
                    setEditContent(comment.content); // 취소 시 원복
                  }}
                  className="text-xs px-2 py-1 text-gray-500 hover:bg-gray-200 rounded flex items-center gap-1"
                  disabled={isLoading}
                >
                  <X size={12} /> 취소
                </button>
                <button
                  onClick={handleUpdate}
                  className="text-xs px-2 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 flex items-center gap-1"
                  disabled={isLoading}
                >
                  <Check size={12} /> 저장
                </button>
              </div>
            </div>
          ) : (
            <p className="text-sm break-words text-gray-700 whitespace-pre-wrap leading-snug">
              {comment.content}
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

// =============================================================================
// [Main Component] 댓글 리스트
// =============================================================================
interface CommentListProps {
  comments: CommentResponse[];
  members: WorkspaceMemberResponse[];
  currentUserId: string; // 💡 추가됨: 권한 체크용
  onRefresh: () => void; // 💡 추가됨: 삭제/수정 시 재조회 요청용
}

const CommentList = ({ comments, members, currentUserId, onRefresh }: CommentListProps) => {
  // 훅 사용
  const { getNickname, getProfileUrl } = useUserLookup(members);
  return (
    <div className="space-y-3 mb-4 max-h-80 overflow-y-auto pr-1 custom-scrollbar">
      {comments.length === 0 ? (
        <p className="text-center text-gray-400 text-sm py-4">작성된 댓글이 없습니다.</p>
      ) : (
        comments.map((comment) => (
          <CommentItem
            key={comment.commentId}
            comment={comment}
            nickname={getNickname(comment.userId)}
            profileUrl={getProfileUrl(comment.userId)}
            currentUserId={currentUserId}
            onRefresh={onRefresh}
          />
        ))
      )}
    </div>
  );
};

export default CommentList;
