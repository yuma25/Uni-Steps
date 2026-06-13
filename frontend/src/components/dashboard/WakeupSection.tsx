import React from 'react';
import { Sunrise, Sun, Edit, Trash2, Users, Moon, Clock, CheckCircle2 } from 'lucide-react';
import type { WakeupCheck, Group } from '../../types';

interface WakeupSectionProps {
  activeWakeup: WakeupCheck | null;
  groupWakeups: WakeupCheck[];
  group: Group | null;
  userId: string;
  onSetWakeup: () => void;
  onEditWakeup: () => void;
  onCancelWakeup: () => void;
  onCheckIn: () => void;
}

const WakeupSection: React.FC<WakeupSectionProps> = ({
  activeWakeup,
  groupWakeups,
  group,
  userId,
  onSetWakeup,
  onEditWakeup,
  onCancelWakeup,
  onCheckIn
}) => {
  return (
    <>
      <section style={{marginBottom: '3rem'}}>
        {activeWakeup ? (
          <div className="wakeup-hero">
            <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
              <div>
                <h3 style={{display: 'flex', alignItems: 'center', gap: '8px', margin: '0 0 0.5rem 0'}}>
                  <Sunrise size={20} /> 現在見守り中
                </h3>
                <p style={{fontSize: '1.5rem', fontWeight: 800, margin: 0}}>
                  {new Date(activeWakeup.target_time).toLocaleString('ja-JP', { hour: '2-digit', minute: '2-digit' })}
                </p>
                <p style={{fontSize: '0.85rem', opacity: 0.9, marginTop: '4px'}}>
                  猶予期間: {activeWakeup.grace_minutes} 分
                </p>
              </div>
              <div style={{display: 'flex', gap: '10px'}}>
                <button 
                  onClick={onEditWakeup} 
                  className="btn-ghost wakeup-action-btn" 
                  title="予定を変更"
                >
                  <Edit size={18} />
                </button>
                <button 
                  onClick={onCancelWakeup} 
                  className="btn-ghost wakeup-action-btn delete" 
                  title="見守りを削除"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
            <button onClick={onCheckIn} className="checkin-btn-modern">
              <Sun size={20} /> 起きました！
            </button>
          </div>
        ) : (
          <div className="task-card" onClick={onSetWakeup} style={{padding: '2.5rem', textAlign: 'center', cursor: 'pointer', border: '2px dashed #e5e7eb'}}>
            <Sunrise size={32} color="var(--brand)" style={{margin: '0 auto 1rem auto'}} />
            <h3 style={{margin: 0, fontSize: '1.1rem'}}>起床見守りをセット</h3>
            <p style={{margin: '8px 0 0 0', fontSize: '0.9rem', color: 'var(--text-secondary)'}}>
              明日の起床予定を仲間に共有して，自分を追い込みましょう．
            </p>
          </div>
        )}
      </section>

      {/* メンバーの状況セクション（常時表示） */}
      <section>
        <div style={{display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '1.5rem'}}>
          <Users size={20} color="var(--text-secondary)" />
          <h2 style={{margin: 0}}>メンバーの状況</h2>
        </div>
        
        <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1rem'}}>
          {group?.users?.map(member => {
            const isMe = member.id === userId;
            const wakeup = groupWakeups.find(w => w.user_id === member.id);
            
            return (
              <div key={member.id} style={{
                display: 'flex', alignItems: 'center', gap: '1rem', 
                background: 'white', padding: '1.25rem', borderRadius: '20px', 
                border: isMe ? '2px solid var(--brand-light)' : '1px solid #e5e7eb',
                boxShadow: 'var(--shadow-sm)',
                opacity: wakeup ? 1 : 0.7
              }}>
                <div style={{
                  width: '44px', height: '44px', borderRadius: '12px', 
                  background: isMe ? 'var(--brand)' : '#f3f4f6',
                  color: isMe ? 'white' : 'var(--text-secondary)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontWeight: 800, fontSize: '1rem'
                }}>
                  {member.name.charAt(0)}
                </div>
                <div style={{flex: 1}}>
                  <div style={{fontWeight: 700, fontSize: '0.95rem'}}>{member.name}{isMe ? ' (自分)' : ''}</div>
                  <div style={{fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '4px'}}>
                    {wakeup ? (
                      <>
                        <Clock size={12} />
                        {new Date(wakeup.target_time).toLocaleString('ja-JP', { hour: '2-digit', minute: '2-digit' })} 予定
                      </>
                    ) : (
                      <>
                        <Moon size={12} />
                        予定なし
                      </>
                    )}
                  </div>
                </div>
                <div>
                  {wakeup ? (
                    wakeup.status === 'pending' ? (
                      <span style={{color: 'var(--warning)', fontWeight: 800, fontSize: '0.75rem'}}>見守り中</span>
                    ) : wakeup.status === 'confirmed' ? (
                      <span style={{color: 'var(--success)', fontWeight: 800, fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '4px'}}>
                        <CheckCircle2 size={12} /> 起きた！
                      </span>
                    ) : (
                      <span style={{color: 'var(--error)', fontWeight: 800, fontSize: '0.75rem'}}>寝坊！</span>
                    )
                  ) : (
                    <span style={{color: 'var(--text-tertiary)', fontWeight: 600, fontSize: '0.75rem'}}>使用なし</span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </section>
    </>
  );
};

export default WakeupSection;
