import React from 'react';
import { X, AlertCircle, Sunrise, Clock, Info } from 'lucide-react';

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
        <button onClick={onClose} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
        <div style={{marginBottom: '2.5rem'}}>
          <h2 style={{margin: 0, fontWeight: 900}}><Sunrise size={24} style={{verticalAlign: 'bottom', marginRight: '10px', color: 'var(--brand)'}} /> 起床見守りをセット</h2>
          <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '0.6rem'}}>明日の朝，仲間に見守ってもらいましょう．</p>
        </div>
        
        <form onSubmit={onSave}>
          <div className="form-group">
            <label><Clock size={14} /> 起床予定時刻</label>
            <input 
              type="datetime-local" 
              required 
              value={formData.target_time} 
              onChange={e => setFormData({...formData, target_time: e.target.value})} 
            />
          </div>
          
          <div className="form-group">
            <label><Info size={14} /> 猶予時間（分）</label>
            <div style={{display: 'flex', gap: '10px', alignItems: 'center'}}>
              <input 
                type="number" 
                required 
                value={isNaN(formData.grace_minutes) ? '' : formData.grace_minutes} 
                onChange={e => {
                  const val = parseInt(e.target.value);
                  setFormData({...formData, grace_minutes: isNaN(val) ? 0 : val});
                }} 
                placeholder="例：15" 
                style={{flex: 1}}
              />
              <span style={{fontWeight: 700, color: 'var(--text-secondary)'}}>分</span>
            </div>
            <p style={{fontSize: '0.8rem', color: 'var(--text-tertiary)', marginTop: '0.75rem'}}>
              この時間を過ぎるとグループ全員に SOS 通知が送信されます．
            </p>
          </div>
          
          <div style={{
            display: 'flex', gap: '12px', padding: '1.25rem', 
            background: '#fffbeb', borderRadius: '16px', border: '1.5px solid #fde68a',
            marginBottom: '2.5rem'
          }}>
            <AlertCircle size={20} color="#f59e0b" style={{flexShrink: 0, marginTop: '2px'}} />
            <p style={{margin: 0, fontSize: '0.85rem', color: '#92400e', lineHeight: 1.6, fontWeight: 500}}>
              <strong>重要:</strong> 設定した時間までに「起きました！」ボタンを押してください．確認できない場合は LINE 等で自動公開されます．
            </p>
          </div>
          
          <button 
            type="submit" 
            disabled={loading} 
            className="btn btn-primary" 
            style={{width: '100%', padding: '16px', borderRadius: '16px', background: 'var(--brand)', fontSize: '1rem'}}
          >
            {loading ? "予約中..." : "見守りを開始する"}
          </button>
        </form>
      </div>
    </div>
  );
};

export default WakeupModal;
