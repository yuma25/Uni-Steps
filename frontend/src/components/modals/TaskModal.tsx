import React from 'react';
import { X } from 'lucide-react';
import type { Task, Group } from '../../types';

interface TaskModalProps {
  show: boolean;
  onClose: () => void;
  editingTask: Task | null;
  formData: {
    title: string;
    deadline: string;
    recurrence_type: string;
    assignees: string[];
  };
  setFormData: React.Dispatch<React.SetStateAction<{
    title: string;
    deadline: string;
    recurrence_type: string;
    assignees: string[];
  }>>;
  onSave: (e: React.FormEvent) => void;
  group: Group | null;
  loading: boolean;
}

const TaskModal: React.FC<TaskModalProps> = ({
  show,
  onClose,
  editingTask,
  formData,
  setFormData,
  onSave,
  group,
  loading
}) => {
  if (!show) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-content animate-pop">
        <button onClick={onClose} className="close-button"><X size={20} /></button>
        <div className="modal-header-text">
          <h2>{editingTask ? "課題を編集" : "新しい課題"}</h2>
          <p>課題の詳細を入力してください．</p>
        </div>
        <form onSubmit={onSave}>
          <div className="form-group">
            <label>タイトル</label>
            <input 
              type="text" 
              required 
              value={formData.title} 
              onChange={e => setFormData({...formData, title: e.target.value})} 
              disabled={editingTask?.source === 'google_classroom'} 
            />
          </div>
          <div className="form-group">
            <label>期限</label>
            <input 
              type="datetime-local" 
              value={formData.deadline} 
              onChange={e => setFormData({...formData, deadline: e.target.value})} 
            />
          </div>
          <div className="form-group">
            <label>該当者</label>
            <div className="assignee-selector">
              {group?.users?.map(member => (
                <label key={member.id} className="assignee-item">
                  <input 
                    type="checkbox" 
                    checked={formData.assignees.includes(member.id)} 
                    onChange={() => {
                      const isSelected = formData.assignees.includes(member.id);
                      setFormData({ 
                        ...formData, 
                        assignees: isSelected 
                          ? formData.assignees.filter(id => id !== member.id) 
                          : [...formData.assignees, member.id] 
                      });
                    }} 
                    disabled={editingTask?.source === 'google_classroom'} 
                  />
                  <span>{member.name}</span>
                </label>
              ))}
            </div>
          </div>
          <button 
            type="submit" 
            disabled={loading} 
            className="icon-button primary full-width" 
            style={{marginTop: '1rem', justifyContent: 'center'}}
          >
            {editingTask ? "更新する" : "登録する"}
          </button>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
