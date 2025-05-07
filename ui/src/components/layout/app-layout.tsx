import React, { ReactNode } from 'react';
import { Sidebar } from './sidebar';
import { Outlet } from '@tanstack/react-router';

interface AppLayoutProps {
  children?: ReactNode;
}

export const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      
      <main className="flex-1 overflow-auto p-6">
        {children || <Outlet />}
      </main>
    </div>
  );
};