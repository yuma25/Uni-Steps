import React from 'react';
import { X, AlertCircle } from 'lucide-react';

interface WakeupModalProps {
  show: boolean;
  onClose: () => void;
  formData: {
    target_time: string;
    grace_minutes: number;
  };
  setFormData: React.Dispatch<React.SetStateAction<{
    target_time: string;
    grace_minutes: number;
  }>>;
  onSave: (e: React.FormEvent) => void;
  loading: boolean;
}

const WakeupModal: React.FC<WakeupModalProps> = ({
  show,
  onClose,
  formData,
  setFormData,
  onSave,
  loading
}) => {
  if (!show) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-content animate-pop">
        <button onClick={onClose} className="close-button"><X size={20} /></button>
        <div className="modal-header-text">
          <h2>起床見守りをセット</h2>
          <p>もし起きられなかった場合，仲間に SOS 通知が飛びます．</p>
        </div>
        <form onSubmit={onSave}>
          <div className="form-group">
            <label>起床予定時刻</label>
            <input 
              type="datetime-local" 
              required 
              value={formData.target_time} 
              onChange={e => setFormData({...formData, target_time: e.target.value})} 
            />
          </div>
          <div className="form-group">
            <label>猶予時間（分）</label>
            <input 
              type="number" 
              required 
              value={isNaN(formData.grace_minutes) ? '' : formData.grace_minutes} 
              onChange={e => {
                const val = parseInt(e.target.value);
                setFormData({...formData, grace_minutes: isNaN(val) ? 0 : val});
              }} 
              placeholder="例：15" 
            />
          </div>
          <div className="alert-info">
            <AlertCircle size={16} />
            <span>設定時刻から猶予時間を過ぎてもチェックインがない場合，部屋のメンバー全員に通知が飛びます．</span>
          </div>
          <button 
            type="submit" 
            disabled={loading} 
            className="icon-button primary full-width" 
            style={{marginTop: '1.5rem', background: 'var(--warning)', border: 'none'}}
          >
            見守りを開始する
          </button>
        </form>
      </div>
    </div>
  );
};

export default WakeupModal;
