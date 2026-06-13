import React from 'react';
import { Sparkles } from 'lucide-react';
import type { NotificationLog } from '../../types';

interface TimelineSectionProps {
  logs: NotificationLog[];
}

const TimelineSection: React.FC<TimelineSectionProps> = ({ logs }) => {
  return (
    <section className="timeline-section">
      <h2><Sparkles size={20} color="var(--primary)" />AI ログ ＆ 通知履歴</h2>
      {logs.length === 0 ? (
        <p className="empty-state">まだ通知の履歴はありません．</p>
      ) : (
        <div className="timeline-list">
          {logs.map(log => {
            const isSOS = log.type === 'sos';
            const isSummary = log.type === 'summary';
            let label = '🤖 AI REMIND';
            let badgeClass = 'remind';
            if (isSOS) { label = '🚨 SOS ALERT'; badgeClass = 'sos'; }
            if (isSummary) { label = '📅 DAILY SUMMARY'; badgeClass = 'summary'; }

            return (
              <div key={log.id} className="timeline-item">
                <div className="timeline-header">
                  <span className={`timeline-badge ${badgeClass}`}>{label}</span>
                  <span className="timeline-time">
                    {new Date(log.created_at).toLocaleString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
                <p className="timeline-message">{log.message}</p>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
};

export default TimelineSection;
