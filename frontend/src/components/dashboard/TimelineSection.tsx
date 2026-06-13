import React from 'react';
import { Sparkles, MessageSquare, AlertTriangle, FileText } from 'lucide-react';
import type { NotificationLog } from '../../types';

interface TimelineSectionProps {
  logs: NotificationLog[];
}

const TimelineSection: React.FC<TimelineSectionProps> = ({ logs }) => {
  return (
    <section>
      <div style={{display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '1.5rem'}}>
        <Sparkles size={20} color="var(--brand)" />
        <h2 style={{margin: 0}}>通知履歴 ＆ AI ログ</h2>
      </div>
      
      {logs.length === 0 ? (
        <div className="empty-state">
          <MessageSquare size={32} color="var(--text-tertiary)" style={{opacity: 0.5, marginBottom: '1rem'}} />
          <p>まだ通知の履歴はありません．</p>
        </div>
      ) : (
        <div style={{display: 'flex', flexDirection: 'column', gap: '1rem'}}>
          {logs.map(log => {
            const isSOS = log.type === 'sos';
            const isSummary = log.type === 'summary';
            
            let Icon = MessageSquare;
            let badgeClass = 'remind';
            let label = 'AI REMIND';
            
            if (isSOS) { Icon = AlertTriangle; badgeClass = 'sos'; label = 'SOS ALERT'; }
            if (isSummary) { Icon = FileText; badgeClass = 'summary'; label = 'SUMMARY'; }

            return (
              <div key={log.id} className="timeline-item">
                <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem'}}>
                  <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                    <Icon size={14} color={isSOS ? 'var(--error)' : isSummary ? 'var(--success)' : 'var(--brand)'} />
                    <span className={`timeline-badge ${badgeClass}`}>{label}</span>
                  </div>
                  <span style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', fontFamily: 'monospace'}}>
                    {new Date(log.created_at).toLocaleString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <p style={{margin: 0, fontSize: '0.95rem', lineHeight: 1.6, color: 'var(--text-primary)'}}>{log.message}</p>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
};

export default TimelineSection;
