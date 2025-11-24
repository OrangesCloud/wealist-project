import { WorkspaceMemberResponse } from '../../types/user';

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
}

export const AvatarStack: React.FC<AvatarStackProps> = ({ members }) => {
  const displayCount = 3;
  const displayMembers = members?.slice(0, displayCount);
  const remainingCount = members?.length - displayCount;

  return (
    <div className="flex -space-x-1.5 p-1 pr-0 overflow-hidden">
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
