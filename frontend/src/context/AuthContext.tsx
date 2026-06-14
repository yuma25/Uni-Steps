import React, { useState, useEffect, useCallback, type ReactNode } from 'react';
import { useSearchParams } from 'react-router-dom';
import { userApi } from '../api/user';
import type { User } from '../types';
import { AuthContext } from './AuthContextInstance';

/**
 * 認証情報を提供するプロバイダーコンポーネントである．
 */
export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [searchParams] = useSearchParams();
  const userIdFromUrl = searchParams.get('user_id') || '';
  
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async (id: string, isRefresh: boolean = false) => {
    if (!id) {
      setLoading(false);
      return;
    }
    
    try {
      // リフレッシュ時のみ，非同期の隙間を作ってから loading をセットして警告を回避する
      if (isRefresh) {
        setLoading(true);
      }

      const userData = await userApi.getMe(id);
      setUser(userData);
    } catch (err: unknown) {
      if (err instanceof Error) {
        console.error("AuthContext: Failed to fetch user", err.message);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // 初回マウント時は初期値 loading: true を利用し，同期的な setState を行わない
    fetchUser(userIdFromUrl, false);
  }, [userIdFromUrl, fetchUser]);

  const refreshUser = useCallback(async () => {
    await fetchUser(userIdFromUrl, true);
  }, [userIdFromUrl, fetchUser]);

  return (
    <AuthContext.Provider value={{ userId: userIdFromUrl, user, loading, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
};
