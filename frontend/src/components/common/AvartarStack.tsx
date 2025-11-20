// src/components/common/AvatarStack.tsx
import React, { useState, useRef, useEffect } from 'react';
import { WorkspaceMemberResponse } from '../../types/user';
import { MessageCircle, X } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';

interface AvatarStackProps {
  members: WorkspaceMemberResponse[];
  onChatClick?: (member: WorkspaceMemberResponse) => void;
}

export const AvatarStack: React.FC<AvatarStackProps> = ({ members, onChatClick }) => {
  const { theme } = useTheme();
  const [showDropdown, setShowDropdown] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const displayCount = 3;
  const displayMembers = members?.slice(0, displayCount);
  const remainingCount = members?.length - displayCount;

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

  // 온라인 상태 (임시 - 나중에 실제 상태로 연동)
  const isOnline = (userId: string) => {
    // TODO: 실제 온라인 상태 확인 로직
    return Math.random() > 0.5;
  };

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
            {members?.map((member) => (
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
                    {/* Online Status Indicator */}
                    <div
                      className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white ${
                        isOnline(member.userId) ? 'bg-green-500' : 'bg-gray-400'
                      }`}
                    />
                  </div>

                  {/* Name + Role */}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-800 truncate">{member.userName}</p>
                    <p className="text-xs text-gray-500">{member.roleName}</p>
                  </div>
                </div>

                {/* Right: Chat Button */}
                <button
                  onClick={() => {
                    onChatClick?.(member);
                    setShowDropdown(false);
                  }}
                  className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-lg transition"
                >
                  <MessageCircle className="w-3.5 h-3.5" />
                  채팅하기
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

interface AssigneeAvatarStackProps {
  assignees: string | string[];
}

export const AssigneeAvatarStack: React.FC<AssigneeAvatarStackProps> = ({ assignees }) => {
  const assigneeList = Array.isArray(assignees)
    ? assignees
    : (assignees as string)
        .split(',')
        .map((name) => name.trim())
        .filter((name) => name.length > 0);

  const initials = assigneeList?.map((name) => name[0]).filter((i) => i);
  const displayCount = 3;

  if (initials.length === 0) {
    return (
      <div
        className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-gray-200 bg-gray-200 text-gray-700`}
      >
        ?
      </div>
    );
  }

  return (
    <div className="flex -space-x-1 p-1 pr-0 overflow-hidden">
      {initials?.slice(0, displayCount)?.map((initial, index) => (
        <div
          key={index}
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white text-white ${
            index === 0 ? 'bg-indigo-500' : index === 1 ? 'bg-pink-500' : 'bg-green-500'
          }`}
          style={{ zIndex: initials.length - index }}
          title={assigneeList[index]}
        >
          {initial}
        </div>
      ))}
      {initials?.length > displayCount && (
        <div
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ring-1 ring-white bg-gray-400 text-white`}
          style={{ zIndex: 0 }}
          title={`${initials?.length - displayCount}명 외`}
        >
          +{initials?.length - displayCount}
        </div>
      )}
    </div>
  );
};
