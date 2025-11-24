import { useEffect, useRef, useState } from 'react';
import { WorkspaceMemberResponse } from '../../types/user';
import { useTheme } from '../../contexts/ThemeContext';
import { getOnlineUsers } from '../../api/chatService';
import { MessageCircle, X } from 'lucide-react';

// =============================================================================
// Helper Function
// =============================================================================
export const getColorByIndex = (index: number) => {
  const colors = ['bg-indigo-500', 'bg-pink-500', 'bg-green-500', 'bg-purple-500', 'bg-yellow-500'];
  return colors[index % colors.length];
};

// =============================================================================
// 💡 개별 멤버 아바타 컴포넌트 (BoardManageModal에서 재사용을 위해 분리)
// =============================================================================
interface MemberAvatarProps {
  member: WorkspaceMemberResponse;
  index: number;
  size?: 'sm' | 'md'; // sm: 24px (스택용), md: 28px (모달 드롭다운용 - BoardManageModal에서 사용할 크기)
}

export const MemberAvatar: React.FC<MemberAvatarProps> = ({ member, index, size = 'sm' }) => {
  const sizeClasses = size === 'md' ? 'w-7 h-7 text-sm' : 'w-6 h-6 text-xs';

  return (
    <div
      key={member.userId}
      className={`${sizeClasses} rounded-full flex items-center justify-center font-bold ring-1 ring-white overflow-hidden flex-shrink-0`}
      style={{ zIndex: index }}
      title={`${member.userName} (${member.roleName})`}
    >
      {member?.profileImageUrl ? (
        <img
          src={member?.profileImageUrl}
          alt={member?.userName}
          className="w-full h-full object-cover"
        />
      ) : (
        <div
          className={`w-full h-full flex items-center justify-center text-white ${getColorByIndex(
            index,
          )}`}
        >
          {member?.userName[0]}
        </div>
      )}
    </div>
  );
};

interface AvatarStackProps {
  members: WorkspaceMemberResponse[];
  onChatClick?: (member: WorkspaceMemberResponse) => void;
}

export const AvatarStack: React.FC<AvatarStackProps> = ({ members, onChatClick }) => {
  const { theme } = useTheme();
  const [showDropdown, setShowDropdown] = useState(false);
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set());
  const [isLoadingOnline, setIsLoadingOnline] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const displayCount = 3;
  const displayMembers = members?.slice(0, displayCount);
  const remainingCount = members?.length - displayCount;

  // 🔥 온라인 사용자 목록 로드
  useEffect(() => {
    // const loadOnlineUsers = async () => {
    //   setIsLoadingOnline(true);
    //   try {
    //     console.log('🔵 [AvatarStack] 온라인 사용자 로딩 시작...');
    //     const users = await getOnlineUsers();
    //     console.log('✅ [AvatarStack] 온라인 사용자 목록:', users);
    //     setOnlineUsers(new Set(users));
    //   } catch (error) {
    //     console.error('❌ [AvatarStack] Failed to load online users:', error);
    //     setOnlineUsers(new Set()); // 에러 시 빈 Set
    //   } finally {
    //     setIsLoadingOnline(false);
    //   }
    // };

    // 드롭다운 열릴 때만 로드
    if (showDropdown) {
      // loadOnlineUsers();
      // 10초마다 갱신 (드롭다운 열려있을 때만)
      // const interval = setInterval(loadOnlineUsers, 10000);
      // return () => clearInterval(interval);
    }
  }, [showDropdown]);

  // 외부 클릭 감지
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    };

    if (showDropdown) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showDropdown]);

  const getColorByIndex = (index: number) => {
    const colors = [
      'bg-indigo-500',
      'bg-pink-500',
      'bg-green-500',
      'bg-purple-500',
      'bg-yellow-500',
    ];
    return colors[index % colors.length];
  };

  // 🔥 온라인 상태 확인 (디버깅 포함)
  const isOnline = (userId: string) => {
    const online = onlineUsers.has(userId);
    console.log(`🟢 [AvatarStack] ${userId} 온라인 상태:`, online);
    return online;
  };

  // 🔥 현재 사용자 확인
  const currentUserId = localStorage.getItem('userId');

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Avatar Stack Button */}
      <button
        onClick={() => setShowDropdown(!showDropdown)}
        className="flex -space-x-1.5 p-1 pr-0 overflow-hidden hover:opacity-80 transition"
      >
        {displayMembers?.map((member, index) => (
          <div
            key={member.userId}
            className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white overflow-hidden"
            style={{ zIndex: members.length - index }}
            title={`${member.userName} (${member.roleName})`}
          >
            {member?.profileImageUrl ? (
              <img
                src={member?.profileImageUrl}
                alt={member?.userName}
                className="w-full h-full object-cover"
              />
            ) : (
              <div
                className={`w-full h-full flex items-center justify-center text-white ${getColorByIndex(
                  index,
                )}`}
              >
                {member?.userName[0]}
              </div>
            )}
          </div>
        ))}
        {remainingCount > 0 && (
          <div
            className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white"
            style={{ zIndex: 0 }}
          >
            +{remainingCount}
          </div>
        )}
      </button>

      {/* Members Dropdown */}
      {showDropdown && (
        <div
          className={`absolute top-full right-0 mt-2 w-80 ${theme.colors.card} shadow-lg ${theme.effects.borderRadius} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 max-h-96 overflow-y-auto`}
        >
          {/* Header */}
          <div className="flex items-center justify-between p-3 border-b">
            <h3 className="text-sm font-semibold text-gray-800">
              프로젝트 멤버 ({members?.length})
              {isLoadingOnline && <span className="ml-2 text-xs text-gray-400">(로딩 중...)</span>}
            </h3>
            <button
              onClick={() => setShowDropdown(false)}
              className="p-1 hover:bg-gray-100 rounded transition"
            >
              <X className="w-4 h-4 text-gray-500" />
            </button>
          </div>

          {/* Members List */}
          <div className="py-2">
            {members?.map((member) => {
              const isCurrentUser = member.userId === currentUserId;
              const memberOnline = isOnline(member.userId);

              return (
                <div
                  key={member.userId}
                  className="flex items-center justify-between px-3 py-2 hover:bg-gray-50 transition"
                >
                  {/* Left: Avatar + Info */}
                  <div className="flex items-center gap-3 flex-1">
                    {/* Avatar with Online Status */}
                    <div className="relative">
                      {member?.profileImageUrl ? (
                        <img
                          src={member?.profileImageUrl}
                          alt={member?.userName}
                          className="w-10 h-10 rounded-full object-cover"
                        />
                      ) : (
                        <div
                          className={`w-10 h-10 rounded-full flex items-center justify-center text-white font-bold ${getColorByIndex(
                            members.indexOf(member),
                          )}`}
                        >
                          {member?.userName[0]}
                        </div>
                      )}
                      {/* 🔥 Online Status Indicator */}
                      <div
                        className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white transition-colors ${
                          memberOnline ? 'bg-green-500' : 'bg-gray-400'
                        }`}
                        title={memberOnline ? '온라인' : '오프라인'}
                      />
                    </div>

                    {/* Name + Role */}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-800 truncate">
                        {member.userName}
                        {isCurrentUser && <span className="ml-1 text-xs text-gray-400">(나)</span>}
                      </p>
                      <p className="text-xs text-gray-500">{member.roleName}</p>
                    </div>
                  </div>

                  {/* Right: Chat Button */}
                  <button
                    onClick={() => {
                      if (!isCurrentUser) {
                        onChatClick?.(member);
                        setShowDropdown(false);
                      }
                    }}
                    disabled={isCurrentUser}
                    className={`flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-lg transition ${
                      isCurrentUser
                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                        : 'text-blue-600 bg-blue-50 hover:bg-blue-100'
                    }`}
                  >
                    <MessageCircle className="w-3.5 h-3.5" />
                    {isCurrentUser ? '나' : '채팅하기'}
                  </button>
                </div>
              );
            })}
          </div>

          {/* 🔥 디버깅 정보 (개발용, 나중에 삭제) */}
          <div className="p-2 border-t bg-gray-50 text-xs text-gray-500">
            <p>온라인 사용자: {onlineUsers.size}명</p>
            <p className="truncate">IDs: {Array.from(onlineUsers).join(', ') || '없음'}</p>
          </div>
        </div>
      )}
    </div>
  );
};

interface AssigneeAvatarStackProps {
  assignees: string | string[];
  workspaceMembers?: WorkspaceMemberResponse[]; // 💡 추가
}

export const AssigneeAvatarStack: React.FC<AssigneeAvatarStackProps> = ({
  assignees,
  workspaceMembers = [],
}) => {
  // 💡 assignees를 배열로 변환
  const assigneeIds = Array.isArray(assignees) ? assignees : [assignees];

  // 💡 userId로 멤버 찾기
  const assigneeMembers = assigneeIds
    .map((userId) => workspaceMembers.find((m) => m.userId === userId))
    .filter((m): m is WorkspaceMemberResponse => m !== undefined);

  const displayCount = 3;
  const displayMembers = assigneeMembers.slice(0, displayCount);
  const remainingCount = assigneeMembers.length - displayCount;

  // 💡 멤버 정보가 없으면 기본 UI
  if (assigneeMembers.length === 0) {
    return (
      <div className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-gray-200 bg-gray-200 text-gray-700">
        ?
      </div>
    );
  }

  return (
    <div className="flex -space-x-1.5 p-1 pr-0 overflow-hidden">
      {displayMembers.map((member, index) => (
        <MemberAvatar
          key={member.userId}
          member={member}
          index={assigneeMembers.length - index}
          size="sm"
        />
      ))}
      {remainingCount > 0 && (
        <div
          className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white"
          style={{ zIndex: 0 }}
        >
          +{remainingCount}
        </div>
      )}
    </div>
  );
};
