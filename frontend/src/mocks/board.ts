import { CustomImportanceResponse, CustomRoleResponse, CustomStageResponse } from '../types/board';

// ⚠️ 임시 Mock Data: API 호출이 제거되었으므로, 컴포넌트 로직을 유지하기 위해 최소한의 Mock 데이터를 사용합니다.
export const MOCK_STAGES: CustomStageResponse[] = [
  // 💡 UUID 형식으로 변경하여 백엔드 검증 통과
  {
    stageId: '00000000-0000-0000-0000-000000000001',
    label: '대기',
    color: '#F59E0B',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '대기 단계',
    isSystemDefault: true,
  },
  {
    stageId: '00000000-0000-0000-0000-000000000002',
    label: '진행중',
    color: '#3B82F6',
    displayOrder: 2,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '진행 단계',
    isSystemDefault: false,
  },
  {
    stageId: '00000000-0000-0000-0000-000000000003',
    label: '완료',
    color: '#10B981',
    displayOrder: 3,
    fieldId: '00000000-0000-0000-0000-000000000010',
    description: '완료 단계',
    isSystemDefault: false,
  },
];
export const MOCK_ROLES: CustomRoleResponse[] = [
  // 💡 UUID 형식으로 변경하여 백엔드 검증 통과
  {
    roleId: '00000000-0000-0000-0000-000000000004',
    label: '프론트엔드',
    color: '#8B5CF6',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000011',
    description: '프론트 역할',
    isSystemDefault: true,
  },
  {
    roleId: '00000000-0000-0000-0000-000000000005',
    label: '백엔드',
    color: '#EC4899',
    displayOrder: 2,
    fieldId: '00000000-0000-0000-0000-000000000011',
    description: '백엔드 역할',
    isSystemDefault: false,
  },
];
export const MOCK_IMPORTANCES: CustomImportanceResponse[] = [
  // 💡 UUID 형식으로 변경하여 백엔드 검증 통과
  {
    importanceId: '00000000-0000-0000-0000-000000000006',
    label: '높음',
    color: '#F59E0B',
    displayOrder: 1,
    fieldId: '00000000-0000-0000-0000-000000000012',
    description: '높은 중요도',
    level: 5,
    isSystemDefault: false,
  },
  {
    importanceId: '00000000-0000-0000-0000-000000000007',
    label: '낮음',
    color: '#10B981',
    displayOrder: 2,
    fieldId: '00000000-0000-0000-0000-000000000012',
    description: '낮은 중요도',
    level: 1,
    isSystemDefault: true,
  },
];
// ⚠️ 주의: 실제 서비스에서는 이 Mock 데이터를 제거하고 새로운 Field/Option API를 구현해야 합니다.
