import { createBrowserRouter } from 'react-router-dom';
import { HomePage } from '../pages/home';
import { AccountsListPage } from '../pages/accounts/list';
import { AccountDetailPage } from '../pages/accounts/detail';
import { AccountCreatePage } from '../pages/accounts/create';
import { TransferCreatePage } from '../pages/transfers/create';
import { TransferDetailPage } from '../pages/transfers/detail';
import { NotFoundPage } from '../pages/not-found';

export const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/accounts', element: <AccountsListPage /> },
  { path: '/accounts/new', element: <AccountCreatePage /> },
  { path: '/accounts/:id', element: <AccountDetailPage /> },
  { path: '/transfers/new', element: <TransferCreatePage /> },
  { path: '/transfers/:id', element: <TransferDetailPage /> },
  { path: '*', element: <NotFoundPage /> },
]);
