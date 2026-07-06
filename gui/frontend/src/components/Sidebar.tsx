// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React from 'react';

const Sidebar: React.FC = () => {
  return (
    <aside className="w-[280px] h-full flex flex-col border-r border-white/10 bg-slate-800 z-10 shrink-0">
      <div className="h-16 flex items-center px-6 border-b border-white/10">
        <h2 className="text-2xl font-bold tracking-tight text-gradient m-0">Anjal</h2>
      </div>
      
      <nav className="flex-1 p-4 overflow-y-auto">
        <ul className="flex flex-col gap-2">
          <li className="flex items-center px-4 py-3 rounded-lg cursor-pointer transition-all duration-150 bg-blue-500/20 text-slate-50 border border-blue-500/30">
            <span className="mr-3 text-lg">🏠</span>
            <span className="font-medium text-sm">Dashboard</span>
          </li>
          <li className="flex items-center px-4 py-3 rounded-lg cursor-pointer transition-all duration-150 text-slate-400 hover:bg-white/5 hover:text-slate-50">
            <span className="mr-3 text-lg">📦</span>
            <span className="font-medium text-sm">Collections</span>
          </li>
          <li className="flex items-center px-4 py-3 rounded-lg cursor-pointer transition-all duration-150 text-slate-400 hover:bg-white/5 hover:text-slate-50">
            <span className="mr-3 text-lg">⚡</span>
            <span className="font-medium text-sm">Requests</span>
          </li>
          <li className="flex items-center px-4 py-3 rounded-lg cursor-pointer transition-all duration-150 text-slate-400 hover:bg-white/5 hover:text-slate-50">
            <span className="mr-3 text-lg">⚙️</span>
            <span className="font-medium text-sm">Settings</span>
          </li>
        </ul>
      </nav>
      
      <div className="p-4 border-t border-white/10">
        <div className="flex items-center gap-3 p-2 rounded-lg cursor-pointer transition-colors hover:bg-white/5">
          <div className="w-9 h-9 rounded-full bg-blue-500 flex items-center justify-center font-semibold text-white">
            U
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-medium text-slate-50">User Workspace</span>
            <span className="text-xs text-slate-400">Online</span>
          </div>
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;
