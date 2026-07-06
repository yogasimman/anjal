// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React, { useRef, useEffect } from 'react';
import { Menu, X, ArrowRight } from 'lucide-react';
import { models } from '../../../../wailsjs/go/models';

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

interface TopUrlBarProps {
  title: string;
  setTitle: (t: string) => void;
  method: string;
  setMethod: (m: string) => void;
  url: string;
  setUrl: (u: string) => void;
  sidebarOpen: boolean;
  setSidebarOpen: (o: boolean) => void;
  onExecute: () => void;
  isExecuting: boolean;
  activeRequest: models.APIRequest | null;
}

export default function TopUrlBar({ title, setTitle, method, setMethod, url, setUrl, sidebarOpen, setSidebarOpen, onExecute, isExecuting, activeRequest }: TopUrlBarProps) {
  const urlInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (urlInputRef.current) {
      urlInputRef.current.scrollLeft = urlInputRef.current.scrollWidth;
    }
  }, [url]);

  return (
    <header className="border-b border-[#444] flex flex-col sm:flex-row sm:items-stretch bg-[#111111]">
      <div className="flex items-center sm:border-r border-[#444] p-4 sm:p-0">
        <button 
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="p-4 text-[#F5F4F0] hover:bg-[#222] transition-colors border-r border-[#444]"
        >
          {sidebarOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
        </button>
        <select 
          value={method}
          onChange={(e) => setMethod(e.target.value)}
          className={`appearance-none bg-transparent w-20 text-center py-4 text-sm font-sans font-bold focus:outline-none cursor-pointer rounded-none
            ${method === 'GET' ? 'text-[#5BA76B]' : method === 'POST' ? 'text-[#E09A38]' : method === 'PUT' ? 'text-[#3B82F6]' : method === 'PATCH' ? 'text-[#A855F7]' : method === 'DELETE' ? 'text-[#E84E38]' : 'text-[#888]'}
          `}
          style={{ textAlignLast: 'center' }}
        >
          {METHODS.map(m => <option key={m} value={m} className="bg-[#111111] text-[#F5F4F0] font-sans text-base">{m}</option>)}
        </select>
      </div>
      
      <div className="flex flex-col border-t sm:border-t-0 border-[#444] flex-1">
        <div className="border-b border-[#444]">
          <input 
            type="text" 
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full bg-transparent border-none text-[#888] text-xs px-6 py-2 focus:outline-none font-semibold uppercase tracking-widest placeholder:text-[#444]"
            placeholder="Request Name"
          />
        </div>
        <input 
          ref={urlInputRef}
          type="text" 
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          className="w-full bg-transparent border-none text-[#F5F4F0] text-lg px-6 py-2 focus:outline-none font-['JetBrains_Mono'] placeholder:text-[#666]"
          placeholder="https://"
        />
      </div>

      <div className="flex border-t sm:border-t-0 sm:border-l border-[#444]">
        <button 
          onClick={onExecute}
          disabled={isExecuting || !activeRequest}
          className={`bg-[#F5F4F0] hover:bg-[#DCDAD2] text-[#111111] font-bold px-8 py-4 h-full text-sm tracking-widest uppercase flex items-center justify-center gap-3 transition-colors ${(!activeRequest || isExecuting) && 'opacity-50 cursor-not-allowed'}`}
        >
          {isExecuting ? 'Sending...' : 'Execute'} <ArrowRight className="w-4 h-4" />
        </button>
      </div>
    </header>
  );
}
