// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React from 'react';

const Dashboard: React.FC = () => {
  return (
    <div className="flex flex-col gap-8 animate-[fadeIn_0.3s_ease-out]">
      <div className="mb-2">
        <h1 className="text-3xl font-semibold tracking-tight mb-2">Welcome to Anjal</h1>
        <p className="text-lg text-slate-400">Your premium API testing and workspace environment.</p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass-panel p-5 flex items-center gap-4 transition-all duration-150 hover:-translate-y-0.5 hover:shadow-lg hover:border-white/20">
          <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center text-xl shrink-0">📦</div>
          <div className="flex flex-col">
            <span className="text-2xl font-bold text-slate-50">12</span>
            <span className="text-sm text-slate-500 uppercase tracking-wide">Collections</span>
          </div>
        </div>
        <div className="glass-panel p-5 flex items-center gap-4 transition-all duration-150 hover:-translate-y-0.5 hover:shadow-lg hover:border-white/20">
          <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center text-xl shrink-0">⚡</div>
          <div className="flex flex-col">
            <span className="text-2xl font-bold text-slate-50">148</span>
            <span className="text-sm text-slate-500 uppercase tracking-wide">Requests Made</span>
          </div>
        </div>
        <div className="glass-panel p-5 flex items-center gap-4 transition-all duration-150 hover:-translate-y-0.5 hover:shadow-lg hover:border-white/20">
          <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center text-xl shrink-0">⏱️</div>
          <div className="flex flex-col">
            <span className="text-2xl font-bold text-slate-50">1.2s</span>
            <span className="text-sm text-slate-500 uppercase tracking-wide">Avg Response Time</span>
          </div>
        </div>
      </div>

      <div className="glass-panel p-6">
        <h3 className="text-xl font-semibold mb-4">Recent Activity</h3>
        <div className="flex flex-col gap-2">
          <div className="flex items-center px-4 py-3 bg-black/20 rounded-lg transition-colors hover:bg-white/5">
            <span className="text-xs font-bold px-2 py-1 rounded w-12 text-center mr-4 bg-blue-500/20 text-blue-400">GET</span>
            <span className="flex-1 font-mono text-slate-400">/api/v1/users</span>
            <span className="text-sm font-medium mr-6 text-emerald-400">200 OK</span>
            <span className="text-xs text-slate-500 min-w-[80px] text-right">2 mins ago</span>
          </div>
          <div className="flex items-center px-4 py-3 bg-black/20 rounded-lg transition-colors hover:bg-white/5">
            <span className="text-xs font-bold px-2 py-1 rounded w-12 text-center mr-4 bg-emerald-500/20 text-emerald-400">POST</span>
            <span className="flex-1 font-mono text-slate-400">/api/v1/auth/login</span>
            <span className="text-sm font-medium mr-6 text-emerald-400">201 Created</span>
            <span className="text-xs text-slate-500 min-w-[80px] text-right">15 mins ago</span>
          </div>
          <div className="flex items-center px-4 py-3 bg-black/20 rounded-lg transition-colors hover:bg-white/5">
            <span className="text-xs font-bold px-2 py-1 rounded w-12 text-center mr-4 bg-red-500/20 text-red-400">DEL</span>
            <span className="flex-1 font-mono text-slate-400">/api/v1/items/42</span>
            <span className="text-sm font-medium mr-6 text-red-400">404 Not Found</span>
            <span className="text-xs text-slate-500 min-w-[80px] text-right">1 hour ago</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
