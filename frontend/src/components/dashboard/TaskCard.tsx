import React from 'react';
import { Calendar, Globe, GraduationCap } from 'lucide-react';
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
    if (isPast) {
      return myProgress 
        ? { label: "提出遅れ", className: "status-overdue" } 
        : { label: "終了", className: "status-closed" };
    }
    return { label: "未完了", className: "status-pending" };
  };

  const status = getStatus();
  const canEdit = task.source !== 'google_classroom' || !task.is_lms_deadline_set;
  const isLMS = task.source === 'google_classroom';

  return (
    <div className="task-card">
      <div className="task-info">
        <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
          <h3>{task.title}</h3>
          {isLMS ? (
            <GraduationCap size={16} color="var(--brand)" />
          ) : (
            <Globe size={16} color="var(--text-tertiary)" />
          )}
        </div>
        <div className="deadline">
          <Calendar size={14} />
          <span>{formatDate(task.deadline)}</span>
        </div>
        
        <div style={{display: 'flex', flexWrap: 'wrap', gap: '6px', marginTop: '1.25rem'}}>
          {task.user_progress?.map(p => (
            <div 
              key={p.user_id} 
              className={`member-badge ${p.is_completed ? 'completed' : ''}`}
              style={{cursor: 'default'}}
              title={p.user_name}
            >
              <div style={{
                width: '18px', height: '18px', borderRadius: '50%', 
                background: p.is_completed ? 'var(--success)' : 'var(--neutral-300)',
                color: 'white', display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '10px', fontWeight: 800, marginRight: '6px'
              }}>
                {p.user_name.charAt(0)}
              </div>
              {p.user_name}
            </div>
          ))}
        </div>
      </div>
      
      <div className="task-footer">
        <span className={`status-badge ${status.className}`}>{status.label}</span>
        <div style={{display: 'flex', gap: '8px'}}>
          {canEdit && (
            <button onClick={() => onEdit(task)} className="btn btn-ghost" style={{padding: '6px 12px', fontSize: '0.8rem'}}>
              編集
            </button>
          )}
          {task.source === 'manual' && (
            <button 
              onClick={() => onToggleCompletion(task.id)} 
              className={`btn ${status.label === "完了" ? 'btn-ghost' : 'btn-primary'}`}
              style={{padding: '6px 16px', fontSize: '0.8rem', background: status.label === "完了" ? '' : 'var(--brand)'}}
            >
              {status.label === "完了" ? "戻す" : "完了"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskCard;
