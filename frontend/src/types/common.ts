export type IROLES = 'OWNER' | 'ORGANIZER' | 'MEMBER' | 'PENDING' | null | undefined
export type IFieldDefaultType = 'stages' | 'roles' | 'importances';
export type IFieldOption = {
    key: IFieldDefaultType;
    value: string;
}