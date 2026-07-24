import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/shared/api/api-client';
import type { Transfer, CreateTransferRequest } from './model';

export const transferApi = {
  getById: (id: string) => apiClient.get<Transfer>(`/transfers/${id}`).then(r => r.data),
  create: (data: CreateTransferRequest) => apiClient.post<Transfer>('/transfers', data).then(r => r.data),
};

export function useTransfer(id: string) {
  return useQuery({
    queryKey: ['transfers', id],
    queryFn: () => transferApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateTransfer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: transferApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      queryClient.invalidateQueries({ queryKey: ['transfers'] });
    },
  });
}
