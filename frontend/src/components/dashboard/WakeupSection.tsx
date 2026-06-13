import React from 'react';
import { Sunrise, Sun, Edit, Trash2 } from 'lucide-react';
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
      <section className="wakeup-section">
        {activeWakeup ? (
          <div className="wakeup-card active">
            <div className="wakeup-info">
              <Sunrise size={24} className="wakeup-icon" />
              <div>
                <h3>起床見守り中</h3>
                <p>予定時刻: {new Date(activeWakeup.target_time).toLocaleString('ja-JP', { hour: '2-digit', minute: '2-digit' })} (+{activeWakeup.grace_minutes}分猶予)</p>
              </div>
            </div>
            <div className="wakeup-actions" style={{display: 'flex', gap: '0.5rem'}}>
              <button onClick={onEditWakeup} className="icon-button" title="時間を変更" style={{background: 'white', color: 'var(--text-sub)'}}>
                <Edit size={16} />
              </button>
              <button onClick={onCancelWakeup} className="icon-button" title="見守りをやめる" style={{background: 'white', color: 'var(--error)'}}>
                <Trash2 size={16} />
              </button>
              <button onClick={onCheckIn} className="checkin-button">
                <Sun size={18} />
                <span>起きました！</span>
              </button>
            </div>
          </div>
        ) : (
          <div className="wakeup-card empty" onClick={onSetWakeup}>
            <Sunrise size={20} />
            <span>明日の起床時間をセットして，仲間に見守ってもらう</span>
          </div>
        )}
      </section>

      {groupWakeups.length > 0 && (
        <section className="group-wakeup-section" style={{marginBottom: '2rem'}}>
          <h3 style={{fontSize: '0.95rem', color: 'var(--text-sub)', marginBottom: '0.8rem'}}>メンバーの起床予定</h3>
          <div className="member-status-list" style={{display: 'flex', flexDirection: 'column', gap: '0.5rem'}}>
            {groupWakeups.map(w => {
              const member = group?.users?.find(u => u.id === w.user_id);
              if (!member) return null;
              const isMe = member.id === userId;
              return (
                <div key={w.id} style={{
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: '0.5rem', 
                  background: isMe ? 'var(--primary-soft)' : '#f8fafc', 
                  border: isMe ? '1px solid var(--primary-light)' : '1px solid transparent',
                  padding: '0.6rem', 
                  borderRadius: '6px', 
                  fontSize: '0.85rem'
                }}>
                  <span style={{fontWeight: 600}}>{member.name}{isMe ? ' (自分)' : ''}</span>
                  <span style={{color: 'var(--text-sub)'}}>: {new Date(w.target_time).toLocaleString('ja-JP', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' })} (+{w.grace_minutes}分)</span>
                  {w.status === 'pending' && <span style={{marginLeft: 'auto', color: 'var(--warning)', fontWeight: 600}}>見守り中</span>}
                  {w.status === 'alerted' && <span style={{marginLeft: 'auto', color: 'var(--error)', fontWeight: 600}}>寝坊！</span>}
                </div>
              );
            })}
          </div>
        </section>
      )}
    </>
  );
};

export default WakeupSection;
