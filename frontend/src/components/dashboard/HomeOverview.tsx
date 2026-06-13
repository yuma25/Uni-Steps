import React from 'react';
import { Clock, CheckCircle2, User } from 'lucide-react';
import type { Task, Group } from '../../types';

interface HomeOverviewProps {
  tasks: Task[];
  group: Group | null;
  userName: string;
  userId: string; // 本人の ID を受け取る
}

const HomeOverview: React.FC<HomeOverviewProps> = ({ tasks, userName, userId }) => {
  const now = new Date();
  const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);
  
  // 今日の課題（グループ全体）
  const todayTasks = tasks.filter(t => {
    const d = new Date(t.deadline);
    return d > now && d <= todayEnd;
  });

  // 自分の未完了かつ期限内の課題
  const myTasks = tasks.filter(t => {
    const deadlineDate = new Date(t.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const isPast = !isUndetermined && deadlineDate < now;
    const myProgress = t.user_progress?.find(p => p.user_id === userId);
    return myProgress && !myProgress.is_completed && !isPast;
  });

  // 現在取り組むべき課題（全員未完了かつ期限内）
  const activeTasks = tasks.filter(t => {
    const deadlineDate = new Date(t.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const isPast = !isUndetermined && deadlineDate < now;
    const allCompleted = t.user_progress && t.user_progress.length > 0 && t.user_progress.every(p => p.is_completed);
    return !allCompleted && !isPast;
  });

  const hour = now.getHours();
  let greeting = "こんにちは";
  if (hour < 5) greeting = "夜更かし中ですね";
  else if (hour < 11) greeting = "おはようございます";
  else if (hour > 18) greeting = "お疲れ様です";

  return (
    <section className="home-overview">
      <div className="overview-welcome">
        <h2 className="greeting-text">{greeting}，{userName || 'ゲスト'}さん</h2>
      </div>

      <div className="overview-stats-grid">
        <div className="stat-card focus">
          <div className="stat-icon-bg" style={{background: 'var(--brand)', color: 'white'}}>
            <User size={20} />
          </div>
          <div className="stat-info">
            <span className="stat-label">あなたのタスク</span>
            <span className="stat-value">{myTasks.length} <small>件</small></span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon-bg" style={{background: 'var(--brand-soft)', color: 'var(--brand)'}}>
            <Clock size={20} />
          </div>
          <div className="stat-info">
            <span className="stat-label">今日の締め切り</span>
            <span className="stat-value">{todayTasks.length} <small>件</small></span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon-bg" style={{background: '#dcfce7', color: 'var(--success)'}}>
            <CheckCircle2 size={20} />
          </div>
          <div className="stat-info">
            <span className="stat-label">取り組むべき全課題</span>
            <span className="stat-value">{activeTasks.length} <small>件</small></span>
          </div>
        </div>
      </div>
    </section>
  );
};

export default HomeOverview;
