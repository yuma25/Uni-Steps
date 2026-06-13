import { createContext } from 'react';
import type { User } from '../types';

/**
 * 認証情報の型定義である．
 */
export interface AuthContextType {
  userId: string;
  user: User | null;
  loading: boolean;
  refreshUser: () => Promise<void>;
}

/**
 * 認証コンテキストのインスタンスである．
 * Fast Refresh の制約により，コンポーネントとは別ファイルで定義する．
 */
export const AuthContext = createContext<AuthContextType | undefined>(undefined);
