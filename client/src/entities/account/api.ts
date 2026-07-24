import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/shared/api/api-client';
import type { Account, CreateAccountRequest } from './model';

export const accountApi = {
  list: () => apiClient.get<Account[]>('/accounts').then(r => r.data),
  getById: (id: string) => apiClient.get<Account>(`/accounts/${id}`).then(r => r.data),
  create: (data: CreateAccountRequest) => apiClient.post<Account>('/accounts', data).then(r => r.data),
  block: (id: string) => apiClient.post(`/accounts/${id}/block`).then(r => r.data),
  unblock: (id: string) => apiClient.post(`/accounts/${id}/unblock`).then(r => r.data),
  close: (id: string) => apiClient.post(`/accounts/${id}/close`).then(r => r.data),
};

export function useAccounts() {
  return useQuery({
    queryKey: ['accounts'],
    queryFn: accountApi.list,
  });
}

export function useAccount(id: string) {
  return useQuery({
    queryKey: ['accounts', id],
    queryFn: () => accountApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: accountApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useBlockAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: accountApi.block,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useUnblockAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: accountApi.unblock,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}

export function useCloseAccount() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: accountApi.close,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
  });
}
