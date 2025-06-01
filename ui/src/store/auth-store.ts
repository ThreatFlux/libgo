import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '@/types/api';
import { login as apiLogin, logout as apiLogout } from '@/api/client';

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  clearError: () => void;
  // Add a hydrate method to fix potential hydration issues
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      
      login: async (username: string, password: string) => {
        try {
          set({ isLoading: true, error: null });
          
          const response = await apiLogin({ username, password });
          
          // Set token to localStorage directly to ensure it's available for API calls
          localStorage.setItem('token', response.token);
          
          set({
            token: response.token,
            user: response.user,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Login failed',
            isLoading: false,
          });
        }
      },
      
      logout: () => {
        // Clear token from localStorage
        localStorage.removeItem('token');
        apiLogout();
        set({
          token: null,
          user: null,
          isAuthenticated: false,
        });
      },
      
      clearError: () => {
        set({ error: null });
      },
      
      // Method to check token in localStorage and update state if needed
      hydrate: () => {
        const token = localStorage.getItem('token');
        const currentToken = get().token;
        
        if (token && !currentToken) {
          set({
            token,
            isAuthenticated: true,
          });
        } else if (!token && currentToken) {
          set({
            token: null,
            user: null,
            isAuthenticated: false,
          });
        }
      }
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated 
      }),
    }
  )
);