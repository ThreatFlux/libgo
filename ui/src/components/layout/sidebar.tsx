import React, { useState } from 'react';
import { Link, useRouter } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { useTheme } from '@/contexts/theme-context';
import { useAuthStore } from '@/store/auth-store';

// Icons
import { 
  LuLayoutDashboard as LuHome, 
  LuServer, 
  LuDownload, 
  LuSettings, 
  LuLogOut,
  LuChevronLeft,
  LuChevronRight,
  LuMoon,
  LuSun,
  LuUser
} from 'react-icons/lu';

const navItems = [
  { 
    name: 'Dashboard', 
    path: '/', 
    icon: <LuHome className="h-5 w-5" /> 
  },
  { 
    name: 'Virtual Machines', 
    path: '/vms', 
    icon: <LuServer className="h-5 w-5" /> 
  },
  { 
    name: 'Exports', 
    path: '/exports', 
    icon: <LuDownload className="h-5 w-5" /> 
  },
  { 
    name: 'Settings', 
    path: '/settings', 
    icon: <LuSettings className="h-5 w-5" /> 
  },
];

export const Sidebar: React.FC = () => {
  const [expanded, setExpanded] = useState(true);
  const { theme, toggleTheme } = useTheme();
  const { logout, user } = useAuthStore();
  const router = useRouter();

  const toggleSidebar = () => {
    setExpanded(!expanded);
  };

  return (
    <div 
      className={cn(
        "flex flex-col h-screen bg-card border-r transition-all duration-300",
        expanded ? "w-64" : "w-16"
      )}
    >
      {/* Logo */}
      <div className="flex items-center justify-between p-4 border-b">
        {expanded ? (
          <h1 className="text-xl font-bold">LibGo KVM</h1>
        ) : (
          <span className="text-xl font-bold mx-auto">LG</span>
        )}
        <Button 
          variant="ghost" 
          size="icon" 
          onClick={toggleSidebar}
          className="ml-auto"
        >
          {expanded ? <LuChevronLeft /> : <LuChevronRight />}
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 px-2">
        <ul className="space-y-2">
          {navItems.map((item) => {
            const isActive = router.state.location.pathname === item.path;
            
            return (
              <li key={item.path}>
                <Link
                  to={item.path}
                  className={cn(
                    "flex items-center p-2 rounded-md transition-colors",
                    isActive 
                      ? "bg-primary/10 text-primary" 
                      : "hover:bg-muted",
                    expanded ? "justify-start" : "justify-center"
                  )}
                >
                  <span className="flex-shrink-0">{item.icon}</span>
                  {expanded && <span className="ml-3">{item.name}</span>}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* Footer */}
      <div className="p-4 border-t space-y-2">
        {/* User */}
        <div className={cn(
          "flex items-center p-2 text-sm text-muted-foreground",
          expanded ? "justify-start" : "justify-center"
        )}>
          <LuUser className="h-4 w-4 flex-shrink-0" />
          {expanded && <span className="ml-3 truncate">{user?.username || 'User'}</span>}
        </div>

        {/* Theme toggle */}
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={toggleTheme}
          className={cn(
            "w-full",
            expanded ? "justify-start" : "justify-center"
          )}
        >
          {theme === 'dark' 
            ? <LuSun className="h-4 w-4 flex-shrink-0" /> 
            : <LuMoon className="h-4 w-4 flex-shrink-0" />
          }
          {expanded && (
            <span className="ml-3">
              {theme === 'dark' ? 'Light mode' : 'Dark mode'}
            </span>
          )}
        </Button>
        
        {/* Logout */}
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={logout}
          className={cn(
            "w-full text-destructive hover:text-destructive",
            expanded ? "justify-start" : "justify-center"
          )}
        >
          <LuLogOut className="h-4 w-4 flex-shrink-0" />
          {expanded && <span className="ml-3">Logout</span>}
        </Button>
      </div>
    </div>
  );
};