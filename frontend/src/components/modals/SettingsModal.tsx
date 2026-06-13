import React, { useState } from 'react';
import { X, Bell, Layout, MessageCircle, Clock, Hash } from 'lucide-react';

interface SettingsModalProps {
  show: boolean;
  onClose: () => void;
  formData: {
    name: string;
    remind_intervals: number[];
    ai_character: string;
    line_channel_token: string;
    line_group_id: string;
    summary_morning_time: string;
    summary_evening_time: string;
  };
  setFormData: React.Dispatch<React.SetStateAction<{
    name: string;
    remind_intervals: number[];
    ai_character: string;
    line_channel_token: string;
    line_group_id: string;
    summary_morning_time: string;
    summary_evening_time: string;
  }>>;
  onSave: (e: React.FormEvent) => void;
  loading: boolean;
}

const SettingsModal: React.FC<SettingsModalProps> = ({
  show,
  onClose,
  formData,
  setFormData,
  onSave,
  loading
}) => {
  const [newInterval, setNewInterval] = useState<string>('');

  if (!show) return null;

  const addInterval = (e: React.MouseEvent) => {
    e.preventDefault();
    const val = parseInt(newInterval);
    if (formData.remind_intervals.length >= 3) {
      alert("リマインド通知は最大 3 つまで設定できます．");
      return;
    }
    if (!isNaN(val) && val > 0 && !formData.remind_intervals.includes(val)) {
      setFormData({
        ...formData,
        remind_intervals: [...formData.remind_intervals, val].sort((a, b) => b - a)
      });
      setNewInterval('');
    }
  };

  const removeInterval = (val: number) => {
    setFormData({
      ...formData,
      remind_intervals: formData.remind_intervals.filter(i => i !== val)
    });
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content animate-pop">
        <button onClick={onClose} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
        <div style={{marginBottom: '2.5rem'}}>
          <h2 style={{margin: 0, fontWeight: 900}}>部屋のカスタマイズ</h2>
          <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', marginTop: '0.6rem'}}>名称，AI の性格，通知タイミングを最適化しましょう．</p>
        </div>
        
        <form onSubmit={onSave}>
          <div className="form-group">
            <label><Hash size={14} /> 部屋の名前</label>
            <input 
              type="text" 
              required 
              value={formData.name} 
              onChange={e => setFormData({...formData, name: e.target.value})} 
              placeholder="例：ゼミ用，月曜2限"
            />
          </div>

          <div className="form-group">
            <label><Layout size={14} /> AI の性格</label>
            <select 
              value={formData.ai_character} 
              onChange={e => setFormData({...formData, ai_character: e.target.value})} 
            >
              <option value="default">標準アシスタント</option>
              <option value="strict">軍隊の厳しい教官</option>
              <option value="kind">心配性な幼馴染</option>
              <option value="cool">冷徹な仕事人執事</option>
            </select>
          </div>

          <div className="form-group">
            <label><MessageCircle size={14} /> LINE 連携設定</label>
            <div style={{display: 'flex', flexDirection: 'column', gap: '0.75rem'}}>
              <input 
                type="text" 
                value={formData.line_channel_token} 
                onChange={e => setFormData({...formData, line_channel_token: e.target.value})} 
                placeholder="Channel Access Token" 
              />
              <input 
                type="text" 
                value={formData.line_group_id} 
                onChange={e => setFormData({...formData, line_group_id: e.target.value})} 
                placeholder="LINE Group ID (C...)" 
              />
            </div>
          </div>

          <div className="form-group">
            <label><Clock size={14} /> サマリー配信時刻</label>
            <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem'}}>
              <div>
                <span style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', fontWeight: 800}}>朝刊 (今日)</span>
                <input 
                  type="time" 
                  value={formData.summary_morning_time} 
                  onChange={e => setFormData({...formData, summary_morning_time: e.target.value})} 
                />
              </div>
              <div>
                <span style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', fontWeight: 800}}>夕刊 (明日)</span>
                <input 
                  type="time" 
                  value={formData.summary_evening_time} 
                  onChange={e => setFormData({...formData, summary_evening_time: e.target.value})} 
                />
              </div>
            </div>
          </div>

          <div className="form-group">
            <label><Bell size={14} /> 通知のタイミング</label>
            <div style={{display: 'flex', flexWrap: 'wrap', gap: '8px', marginBottom: '1rem'}}>
              {formData.remind_intervals.map(val => (
                <div key={val} className="timeline-badge remind" style={{display: 'flex', alignItems: 'center', gap: '8px', padding: '8px 12px', borderRadius: '12px'}}>
                  <span style={{fontWeight: 800}}>{val >= 1440 ? `${val/1440}日` : val >= 60 ? `${val/60}時間` : `${val}分`}前</span>
                  <button type="button" onClick={() => removeInterval(val)} style={{background: 'none', border: 'none', padding: 0, display: 'flex', cursor: 'pointer', color: 'var(--brand)'}}><X size={14} /></button>
                </div>
              ))}
            </div>
            <div style={{display: 'flex', gap: '10px'}}>
              <input type="number" value={newInterval} onChange={e => setNewInterval(e.target.value)} placeholder="分前を入力" />
              <button type="button" onClick={addInterval} className="btn btn-ghost" style={{whiteSpace: 'nowrap', borderRadius: '16px'}}>追加</button>
            </div>
          </div>

          <button 
            type="submit" 
            disabled={loading} 
            className="btn btn-primary" 
            style={{width: '100%', padding: '16px', marginTop: '1rem', borderRadius: '16px'}}
          >
            {loading ? "更新中..." : "全ての設定を保存する"}
          </button>
        </form>
      </div>
    </div>
  );
};

export default SettingsModal;
