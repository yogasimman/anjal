import React from 'react';

const Header: React.FC = () => {
  return (
    <header className="h-16 flex items-center justify-between px-6 border-b border-white/10 glass-panel !rounded-none sticky top-0 z-50">
      <div className="flex-1 max-w-[400px]">
        <input 
          type="text" 
          placeholder="Search requests..." 
          className="w-full py-2 px-4 bg-black/20 border border-white/10 rounded-full text-slate-50 text-sm outline-none transition-all duration-150 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/50" 
        />
      </div>
      <div className="flex items-center gap-4">
        <button className="p-2 rounded-lg text-slate-400 text-sm font-medium transition-all duration-150 flex items-center justify-center hover:bg-white/10 hover:text-slate-50" title="Notifications">
          🔔
        </button>
        <button className="px-4 py-2 rounded-lg text-sm font-medium transition-all duration-150 flex items-center justify-center bg-blue-500 text-white shadow-sm hover:bg-blue-600 hover:shadow-[0_0_15px_rgba(59,130,246,0.5)]" title="New Request">
          + New Request
        </button>
      </div>
    </header>
  );
};

export default Header;
