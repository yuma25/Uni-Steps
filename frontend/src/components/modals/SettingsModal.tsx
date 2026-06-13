import React, { useState } from 'react';
import { X } from 'lucide-react';

interface SettingsModalProps {
  show: boolean;
  onClose: () => void;
  formData: {
    remind_intervals: number[];
    ai_character: string;
    line_channel_token: string;
    line_group_id: string;
    summary_morning_time: string;
    summary_evening_time: string;
  };
  setFormData: React.Dispatch<React.SetStateAction<{
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
      alert("最大 3 つまでです．");
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
        <button onClick={onClose} className="close-button"><X size={20} /></button>
        <div className="modal-header-text">
          <h2>部屋の設定</h2>
          <p>通知タイミングや AI の性格を変更できます．</p>
        </div>
        <form onSubmit={onSave}>
          <div className="form-group">
            <label>AI のキャラクター設定</label>
            <select 
              className="full-width" 
              value={formData.ai_character} 
              onChange={e => setFormData({...formData, ai_character: e.target.value})} 
              style={{padding: '0.8rem', borderRadius: '8px', border: '1px solid var(--neutral-300)'}}
            >
              <option value="default">標準アシスタント</option>
              <option value="strict">厳しい教官</option>
              <option value="kind">心配性な幼馴染</option>
              <option value="cool">冷徹な執事</option>
            </select>
          </div>
          <div className="form-group">
            <label>LINE Bot 連携（オプション）</label>
            <input 
              type="text" 
              className="full-width" 
              style={{marginBottom: '0.5rem'}} 
              value={formData.line_channel_token} 
              onChange={e => setFormData({...formData, line_channel_token: e.target.value})} 
              placeholder="LINE Channel Token" 
            />
            <input 
              type="text" 
              className="full-width" 
              value={formData.line_group_id} 
              onChange={e => setFormData({...formData, line_group_id: e.target.value})} 
              placeholder="LINE Group ID (例: Cxxxxx...)" 
            />
            <p style={{fontSize: '0.8rem', color: 'var(--text-sub)', margin: '0.5rem 0 0 0'}}>
              ※設定すると，SOS 通知などが指定した LINE グループにも送信されます．
            </p>
          </div>
          <div className="form-group" style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem'}}>
            <div>
              <label>朝のサマリー時刻</label>
              <input 
                type="time" 
                className="full-width" 
                value={formData.summary_morning_time} 
                onChange={e => setFormData({...formData, summary_morning_time: e.target.value})} 
              />
            </div>
            <div>
              <label>夜のサマリー時刻</label>
              <input 
                type="time" 
                className="full-width" 
                value={formData.summary_evening_time} 
                onChange={e => setFormData({...formData, summary_evening_time: e.target.value})} 
              />
            </div>
          </div>
          <div className="form-group">
            <label>リマインド通知タイミング（分前 / 最大3つ）</label>
            <div className="interval-list">
              {formData.remind_intervals.map(val => (
                <div key={val} className="interval-tag">
                  <span>{val >= 1440 ? `${val/1440}日` : val >= 60 ? `${val/60}時間` : `${val}分`}前</span>
                  <button type="button" onClick={() => removeInterval(val)} className="remove-btn"><X size={12} /></button>
                </div>
              ))}
            </div>
            <div className="interval-input-group">
              <input type="number" value={newInterval} onChange={e => setNewInterval(e.target.value)} placeholder="例：30" />
              <button type="button" onClick={addInterval} className="icon-button">追加</button>
            </div>
          </div>
          <button 
            type="submit" 
            disabled={loading} 
            className="icon-button primary full-width" 
            style={{marginTop: '1rem', justifyContent: 'center'}}
          >
            設定を保存する
          </button>
        </form>
      </div>
    </div>
  );
};

export default SettingsModal;
