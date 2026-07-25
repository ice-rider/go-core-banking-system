export interface Transfer {
  id: string;
  from_account_id: string;
  to_account_id: string;
  amount: number;
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
  idempotency_key: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransferRequest {
  from_account_id: string;
  to_account_id: string;
  amount: number;
  idempotency_key: string;
}
