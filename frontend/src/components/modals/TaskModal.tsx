import React from 'react';
import { X, Calendar, CheckCircle2, User, AlertCircle, ClipboardList, Trash2 } from 'lucide-react';
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
  onDelete: (taskId: string) => void;
  group: Group | null;
  loading: boolean;
  operatorId: string;
  ownerId: string;
}

const TaskModal: React.FC<TaskModalProps> = ({
  show,
  onClose,
  editingTask,
  formData,
  setFormData,
  onSave,
  onDelete,
  group,
  loading,
  operatorId,
  ownerId
}) => {
  if (!show) return null;

  // 権限チェック：作成者 または グループオーナーのみがタイトル・期限・削除の操作が可能
  const isAuthorized = !editingTask || editingTask.creator_id === operatorId || ownerId === operatorId;
  const isLMS = editingTask?.source === 'google_classroom';

  // 担当者が一人も選択されていない場合は保存できないようにする．
  const isAssigneeEmpty = formData.assignees.length === 0;

  return (
    <div className="modal-overlay">
      <div className="modal-content animate-pop">
        <button onClick={onClose} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
        <div className="modal-header-text" style={{marginBottom: '2.5rem'}}>
          <h2 style={{margin: 0, fontWeight: 900}}>{editingTask ? "課題を編集" : "新しい課題を登録"}</h2>
          <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '0.6rem'}}>詳細を入力して，チームと共有しましょう．</p>
        </div>
        
        <form onSubmit={onSave}>
          <div className="form-group">
            <label><ClipboardList size={14} /> タイトル</label>
            <input 
              type="text" 
              required 
              placeholder="例：レポート提出，プレゼン準備"
              value={formData.title} 
              onChange={e => setFormData({...formData, title: e.target.value})} 
              disabled={isLMS || !isAuthorized} 
            />
            {!isAuthorized && <p style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.4rem'}}>※タイトルの変更は作成者またはオーナーのみ可能です．</p>}
          </div>
          
          <div className="form-group">
            <label><Calendar size={14} /> 期限</label>
            <input 
              type="datetime-local" 
              value={formData.deadline} 
              onChange={e => setFormData({...formData, deadline: e.target.value})} 
              disabled={isLMS || !isAuthorized}
            />
            {!isAuthorized && <p style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.4rem'}}>※期限の変更は作成者またはオーナーのみ可能です．</p>}
          </div>
          
          <div className="form-group">
            <label><User size={14} /> 担当者を編集</label>
            <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '0.75rem', marginTop: '0.5rem'}}>
              {group?.users?.map(member => {
                const isMe = member.id === operatorId;
                const canToggle = isAuthorized || isMe;
                
                return (
                  <button
                    key={member.id}
                    type="button"
                    onClick={() => {
                      if (isLMS || !canToggle) return;
                      const isSelected = formData.assignees.includes(member.id);
                      setFormData({ 
                        ...formData, 
                        assignees: isSelected 
                          ? formData.assignees.filter(id => id !== member.id) 
                          : [...formData.assignees, member.id] 
                      });
                    }}
                    style={{
                      display: 'flex', alignItems: 'center', gap: '10px', padding: '12px',
                      background: formData.assignees.includes(member.id) ? 'var(--brand-soft)' : '#f8fafc',
                      border: `2px solid ${formData.assignees.includes(member.id) ? 'var(--brand-light)' : '#f1f5f9'}`,
                      borderRadius: '14px', 
                      cursor: (isLMS || !canToggle) ? 'not-allowed' : 'pointer',
                      opacity: canToggle ? 1 : 0.6,
                      transition: 'all 0.2s', textAlign: 'left'
                    }}
                  >
                    <div style={{
                      width: '24px', height: '24px', borderRadius: '50%',
                      background: formData.assignees.includes(member.id) ? 'var(--brand)' : 'var(--text-tertiary)',
                      display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: '10px', fontWeight: 900
                    }}>
                      {formData.assignees.includes(member.id) ? <CheckCircle2 size={14} /> : member.name.charAt(0)}
                    </div>
                    <span style={{fontWeight: 700, fontSize: '0.9rem', color: formData.assignees.includes(member.id) ? 'var(--brand)' : 'var(--text-secondary)'}}>{member.name}</span>
                  </button>
                );
              })}
            </div>
            {!isAuthorized && <p style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '0.8rem'}}>※作成者・オーナー以外は，自分自身の参加・辞退のみ操作可能です．</p>}
            {isAssigneeEmpty && (
              <div style={{display: 'flex', alignItems: 'center', gap: '6px', marginTop: '0.75rem', color: 'var(--warning)', fontSize: '0.85rem', fontWeight: 600}}>
                <AlertCircle size={14} />
                <span>担当者がいないため，保存するとこの課題は削除されます．</span>
              </div>
            )}
            </div>

            <div style={{marginTop: '2.5rem', display: 'flex', flexDirection: 'column', gap: '1rem'}}>
            <button 
              type="submit" 
              disabled={loading} 
              className="btn btn-primary" 
              style={{width: '100%', padding: '16px', marginTop: '1.5rem', borderRadius: '16px'}}
            >
              {loading ? "保存中..." : (editingTask ? "更新を保存する" : "課題を登録する")}
            </button>

            {editingTask && isAuthorized && !isLMS && (
              <button
                type="button"
                onClick={() => onDelete(editingTask.id)}
                disabled={loading}
                className="btn btn-ghost"
                style={{width: '100%', padding: '14px', borderRadius: '16px', color: 'var(--error)', borderColor: '#fee2e2'}}
              >
                <Trash2 size={18} />
                課題を削除する
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
