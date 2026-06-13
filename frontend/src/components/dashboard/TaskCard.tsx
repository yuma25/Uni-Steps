import React from 'react';
import { Calendar } from 'lucide-react';
import type { Task } from '../../types';
import { formatDate } from '../../utils/helpers';

interface TaskCardProps {
  task: Task;
  userId: string;
  onEdit: (task: Task) => void;
  onToggleCompletion: (taskId: string) => void;
}

const TaskCard: React.FC<TaskCardProps> = ({ task, userId, onEdit, onToggleCompletion }) => {
  const myProgress = task.user_progress?.find(p => p.user_id === userId);
  const deadlineDate = new Date(task.deadline);
  const isUndetermined = deadlineDate.getFullYear() <= 1;
  const isPast = !isUndetermined && deadlineDate < new Date();

  const getStatus = () => {
    if (myProgress?.is_completed) return { label: "完了", className: "status-completed" };
    if (isPast) return { label: "提出遅れ", className: "status-overdue" };
    return { label: "未完了", className: "status-pending" };
  };

  const status = getStatus();
  const canEdit = task.source !== 'google_classroom' || !task.is_lms_deadline_set;

  return (
    <div className="task-card">
      <div className="task-info">
        <h3>{task.title}</h3>
        <p className="deadline">
          <Calendar size={13} />
          <span>{formatDate(task.deadline)}</span>
        </p>
        <div className="member-status-list">
          {task.user_progress?.map(p => (
            <span key={p.user_id} className={`member-badge ${p.is_completed ? 'completed' : ''}`}>
              {p.user_name || "User"}: {p.is_completed ? "完了" : "未"}
            </span>
          ))}
        </div>
      </div>
      <div className="task-status">
        <span className={`status-badge ${status.className}`}>{status.label}</span>
        <div className="card-actions">
          {canEdit && (
            <button onClick={() => onEdit(task)} className="edit-button">編集</button>
          )}
          {task.source !== 'google_classroom' && (
            <button 
              onClick={() => onToggleCompletion(task.id)} 
              className="complete-toggle-button"
            >
              {status.label === "完了" ? "未完了に戻す" : "完了にする"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskCard;
