import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { useCreateAccount } from '@/entities/account/api';

const createAccountSchema = z.object({
  owner_name: z.string().min(1, 'Owner name is required'),
});

type CreateAccountFormData = z.infer<typeof createAccountSchema>;

export function CreateAccountForm() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const createAccount = useCreateAccount();

  const { register, handleSubmit, formState: { errors } } = useForm<CreateAccountFormData>({
    resolver: zodResolver(createAccountSchema),
  });

  const onSubmit = async (data: CreateAccountFormData) => {
    await createAccount.mutateAsync(data);
    navigate('/accounts');
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="owner_name">{t('accounts.owner_name')}</Label>
        <Input id="owner_name" {...register('owner_name')} />
        {errors.owner_name && (
          <p className="text-sm text-destructive">{errors.owner_name.message}</p>
        )}
      </div>
      <Button type="submit" disabled={createAccount.isPending}>
        {createAccount.isPending ? t('common.loading') : t('common.save')}
      </Button>
    </form>
  );
}
