import React from 'react';
import { Plus } from 'lucide-react';

interface FABProps {
  onClick: () => void;
}

const FAB: React.FC<FABProps> = ({ onClick }) => {
  return (
    <button className="fab" onClick={onClick} title="課題を追加">
      <Plus size={32} />
    </button>
  );
};

export default FAB;
