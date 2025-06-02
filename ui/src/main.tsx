import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@/contexts/theme-context';
import {
  Outlet,
  RouterProvider,
  Router,
  redirect,
  createRootRoute,
  createRoute,
} from '@tanstack/react-router';
import { useAuthStore } from '@/store/auth-store';
import { AppLayout } from '@/components/layout/app-layout';
import { DashboardPage } from '@/pages/dashboard';
import { LoginPage } from '@/pages/login';
import { VMListPage } from '@/pages/vm-list';
import { VMDetailPage } from '@/pages/vm-detail';
import { VMCreatePage } from '@/pages/vm-create';
import { VMExportPage } from '@/pages/vm-export';
import { ExportsPage } from '@/pages/exports';
import { NetworkList } from '@/pages/network-list';
import { NetworkCreate } from '@/pages/network-create';
import { NetworkDetail } from '@/pages/network-detail';
import { StorageList } from '@/pages/storage-list';
import { StorageCreate } from '@/pages/storage-create';
import { StorageDetail } from '@/pages/storage-detail';
import '@/styles/globals.css';

// Create a query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000, // 1 minute
    },
  },
});

// Create a root route
const rootRoute = createRootRoute({
  component: () => <Outlet />,
});

// Auth check function
const checkAuth = () => {
  // Hydrate auth store from localStorage to ensure state is up-to-date
  useAuthStore.getState().hydrate();
  
  const { isAuthenticated } = useAuthStore.getState();
  if (!isAuthenticated) {
    throw redirect({
      to: '/login',
    });
  }
};

// Login route
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
  beforeLoad: () => {
    // Hydrate auth store from localStorage to ensure state is up-to-date
    useAuthStore.getState().hydrate();
    
    const { isAuthenticated } = useAuthStore.getState();
    if (isAuthenticated) {
      throw redirect({
        to: '/',
      });
    }
  },
});

// App layout route
const appLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'app',
  component: AppLayout,
  beforeLoad: checkAuth,
});

// Dashboard route
const dashboardRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/',
  component: DashboardPage,
});

// VM list route
const vmListRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/vms',
  component: VMListPage,
});

// VM create route
const vmCreateRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/vms/create',
  component: VMCreatePage,
});

// VM detail route
const vmDetailRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/vms/$name',
  component: VMDetailPage,
});

// VM export route
const vmExportRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/vms/$name/export',
  component: VMExportPage,
});

// Exports route
const exportsRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/exports',
  component: ExportsPage,
});

// Network list route
const networkListRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/networks',
  component: NetworkList,
});

// Network create route
const networkCreateRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/networks/create',
  component: NetworkCreate,
});

// Network detail route
const networkDetailRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/networks/$name',
  component: NetworkDetail,
});

// Storage list route
const storageListRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/storage',
  component: StorageList,
});

// Storage create route
const storageCreateRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/storage/create',
  component: StorageCreate,
});

// Storage detail route
const storageDetailRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/storage/pools/$poolName',
  component: StorageDetail,
});

// Create the router
const router = new Router({
  routeTree: rootRoute.addChildren([
    loginRoute,
    appLayoutRoute.addChildren([
      dashboardRoute,
      vmListRoute,
      vmCreateRoute,
      vmDetailRoute,
      vmExportRoute,
      exportsRoute,
      networkListRoute,
      networkCreateRoute,
      networkDetailRoute,
      storageListRoute,
      storageCreateRoute,
      storageDetailRoute,
    ]),
  ]),
  defaultPreload: 'intent',
});

// Render the app
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <RouterProvider router={router} />
      </ThemeProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);