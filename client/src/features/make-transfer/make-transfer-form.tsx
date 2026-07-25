import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select';
import { accountApi } from '@/entities/account/api';
import { useCreateTransfer } from '@/entities/transfer/api';

const createTransferSchema = z.object({
  from_account_id: z.string().min(1, 'Source account is required'),
  to_account_id: z.string().min(1, 'Destination account is required'),
  amount: z.number().min(1, 'Amount must be greater than 0'),
  idempotency_key: z.string().min(1, 'Idempotency key is required'),
});

type CreateTransferFormData = z.infer<typeof createTransferSchema>;

export function MakeTransferForm() {
  const { t } = useTranslation();
  const createTransfer = useCreateTransfer();

  const { data: accounts = [] } = useQuery({
    queryKey: ['accounts'],
    queryFn: accountApi.list,
  });

  const { register, handleSubmit, setValue, formState: { errors } } = useForm<CreateTransferFormData>({
    resolver: zodResolver(createTransferSchema),
    defaultValues: {
      idempotency_key: crypto.randomUUID(),
    },
  });

  const onSubmit = async (data: CreateTransferFormData) => {
    await createTransfer.mutateAsync(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label>{t('transfers.from')}</Label>
        <Select onValueChange={(value: string) => setValue('from_account_id', value)}>
          <SelectTrigger>
            <SelectValue placeholder={t('transfers.from')} />
          </SelectTrigger>
          <SelectContent>
            {accounts.map((account) => (
              <SelectItem key={account.id} value={account.id}>
                {account.owner_name} ({account.id})
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {errors.from_account_id && (
          <p className="text-sm text-destructive">{errors.from_account_id.message}</p>
        )}
      </div>
      <div className="space-y-2">
        <Label>{t('transfers.to')}</Label>
        <Select onValueChange={(value: string) => setValue('to_account_id', value)}>
          <SelectTrigger>
            <SelectValue placeholder={t('transfers.to')} />
          </SelectTrigger>
          <SelectContent>
            {accounts.map((account) => (
              <SelectItem key={account.id} value={account.id}>
                {account.owner_name} ({account.id})
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {errors.to_account_id && (
          <p className="text-sm text-destructive">{errors.to_account_id.message}</p>
        )}
      </div>
      <div className="space-y-2">
        <Label>{t('transfers.amount')}</Label>
        <Input type="number" {...register('amount', { valueAsNumber: true })} />
        {errors.amount && (
          <p className="text-sm text-destructive">{errors.amount.message}</p>
        )}
      </div>
      <input type="hidden" {...register('idempotency_key')} />
      <Button type="submit" disabled={createTransfer.isPending}>
        {createTransfer.isPending ? t('common.loading') : t('transfers.submit')}
      </Button>
    </form>
  );
}
