export interface Account {
  id: string;
  owner_name: string;
  balance: number;
  status: 'ACTIVE' | 'BLOCKED' | 'CLOSED';
  created_at: string;
  updated_at: string;
}

export interface CreateAccountRequest {
  owner_name: string;
}
