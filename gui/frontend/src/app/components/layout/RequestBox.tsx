import React, { useState } from 'react';
import { models } from '../../../../wailsjs/go/models';

interface RequestBoxProps {
  activeRequest: models.APIRequest | null;
}

export default function RequestBox({ activeRequest }: RequestBoxProps) {
  const [activeTab, setActiveTab] = useState('Params');

  return (
    <div className="flex-1 flex flex-col border-b lg:border-b-0 lg:border-r border-[#444] min-h-0 bg-[#111111]">
      <div className="flex border-b border-[#444]">
        {['Params', 'Headers', 'Body', 'Auth'].map(tab => (
          <button 
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-6 py-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444] transition-colors
              ${activeTab === tab 
                ? 'bg-[#F5F4F0] text-[#111111]' 
                : 'text-[#F5F4F0] hover:bg-[#222]'}
            `}
          >
            {tab}
          </button>
        ))}
      </div>
      
      <div className="flex-1 overflow-auto p-0">
        {activeTab === 'Params' && <KeyValueEditor data={activeRequest?.QueryParams} />}
        {activeTab === 'Headers' && <KeyValueEditor data={activeRequest?.Headers} headers />}
        {activeTab === 'Body' && (
          <div className="h-full flex flex-col p-6">
            <div className="flex items-center gap-6 mb-6 text-xs uppercase tracking-widest font-semibold text-[#888]">
              <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
                <input type="radio" name="bodyType" className="accent-[#F5F4F0]" /> None
              </label>
              <label className="flex items-center gap-2 cursor-pointer text-[#F5F4F0]">
                <input type="radio" name="bodyType" defaultChecked className="accent-[#F5F4F0]" /> Raw / JSON
              </label>
            </div>
            <div className="flex-1 border border-[#444] p-6 font-['JetBrains_Mono'] text-sm overflow-auto text-[#F5F4F0] bg-[#1A1A1A] relative shadow-[4px_4px_0px_0px_rgba(68,68,68,1)]">
              <pre className="outline-none min-h-full whitespace-pre-wrap">{activeRequest?.Body || ''}</pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function KeyValueEditor({ headers = false, data = {} }: { headers?: boolean, data?: Record<string, string> }) {
  const entries = Object.entries(data || {});
  const emptyRows = Math.max(3 - entries.length, 1);
  
  return (
    <div className="flex flex-col">
      <div className="grid grid-cols-12 border-b border-[#444] bg-[#222]">
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">Key</div>
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">Value</div>
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold">Description</div>
      </div>
      
      {entries.map(([k, v], i) => (
        <div key={i} className="grid grid-cols-12 border-b border-[#444] group bg-[#111111]">
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" value={k} readOnly />
          </div>
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" value={v} readOnly />
          </div>
          <div className="col-span-4 relative flex items-center">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] text-[#888] focus:outline-none focus:bg-[#1A1A1A]" readOnly />
          </div>
        </div>
      ))}
      
      {[...Array(emptyRows)].map((_, i) => (
        <div key={`empty-${i}`} className="grid grid-cols-12 border-b border-[#444] last:border-b-0 group bg-[#111111]">
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" placeholder="Key..." />
          </div>
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" placeholder="Value..." />
          </div>
          <div className="col-span-4 relative flex items-center">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] text-[#888] focus:outline-none focus:bg-[#1A1A1A]" placeholder="Description..." />
          </div>
        </div>
      ))}
    </div>
  );
}
