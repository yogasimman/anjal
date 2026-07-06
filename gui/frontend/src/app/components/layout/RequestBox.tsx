// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React, { useState, useRef } from 'react';
import { models } from '../../../../wailsjs/go/models';

interface RequestBoxProps {
  body: string;
  setBody: (b: string) => void;
  headers: Record<string, string>;
  setHeaders: (h: Record<string, string>) => void;
  params: Record<string, string>;
  setParams: (p: Record<string, string>) => void;
  auth: models.Auth | undefined;
  setAuth: (a: models.Auth | undefined) => void;
}

export default function RequestBox({ body, setBody, headers, setHeaders, params, setParams, auth, setAuth }: RequestBoxProps) {
  const [activeTab, setActiveTab] = useState('Params');

  return (
    <div className="flex-1 flex flex-col border-b lg:border-b-0 lg:border-r border-[#444] min-h-0 bg-[#111111]">
      <div className="flex border-b border-[#444]">
        {['Params', 'Headers', 'Auth', 'Body'].map(tab => (
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
        {activeTab === 'Params' && <KeyValueEditor key="params" data={params} onChange={setParams} />}
        {activeTab === 'Headers' && <KeyValueEditor key="headers" data={headers} onChange={setHeaders} headers />}
        {activeTab === 'Auth' && <AuthEditor auth={auth} setAuth={setAuth} />}
        {activeTab === 'Body' && (
          <div className="h-full flex flex-col p-6">
            <div className="flex items-center gap-6 mb-6 text-xs uppercase tracking-widest font-semibold text-[#888]">
              <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
                <input type="radio" name="bodyType" className="accent-[#F5F4F0]" /> None
              </label>
              <label className="flex items-center gap-2 cursor-pointer text-[#F5F4F0] hover:text-white transition-colors">
                <input type="radio" name="bodyType" defaultChecked className="accent-[#F5F4F0]" /> JSON
              </label>
              <label className="flex items-center gap-2 cursor-pointer text-[#888] hover:text-white transition-colors">
                <input type="radio" name="bodyType" className="accent-[#F5F4F0]" /> Raw
              </label>
            </div>
            <div className="flex-1 border border-[#444] p-0 font-['JetBrains_Mono'] text-sm overflow-hidden text-[#F5F4F0] bg-[#1A1A1A] relative shadow-[4px_4px_0px_0px_rgba(68,68,68,1)]">
              <textarea 
                className="w-full h-full p-6 bg-transparent outline-none resize-none whitespace-pre-wrap"
                value={body}
                onChange={e => setBody(e.target.value)}
                placeholder="Enter request body here..."
                spellCheck={false}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function KeyValueEditor({ headers = false, data = {}, onChange }: { headers?: boolean, data?: Record<string, string>, onChange: (d: Record<string, string>) => void }) {
  const initialEntries = Object.entries(data || {}).map(([key, value]) => ({ key, value }));
  initialEntries.push({ key: '', value: '' });
  
  const [entries, setEntries] = useState(initialEntries);
  const [colWidths, setColWidths] = useState([250, 250]);
  const dragInfo = useRef({ isDragging: false, colIndex: 0, startX: 0, startWidth: 0 });

  const handleMouseDown = (e: React.MouseEvent, colIndex: number) => {
    e.preventDefault();
    dragInfo.current = { isDragging: true, colIndex, startX: e.clientX, startWidth: colWidths[colIndex] };
    
    const handleMouseMove = (moveEvent: MouseEvent) => {
      if (!dragInfo.current.isDragging) return;
      const delta = moveEvent.clientX - dragInfo.current.startX;
      setColWidths(prev => {
        const newWidths = [...prev];
        newWidths[dragInfo.current.colIndex] = Math.max(100, dragInfo.current.startWidth + delta);
        return newWidths;
      });
    };

    const handleMouseUp = () => {
      dragInfo.current.isDragging = false;
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  };

  const gridStyle = {
    display: 'grid',
    gridTemplateColumns: `${colWidths[0]}px ${colWidths[1]}px minmax(150px, 1fr)`
  };

  const handleChange = (index: number, key: string, value: string) => {
    const newEntries = [...entries];
    newEntries[index] = { key, value };
    
    // Add new empty row if we just typed in the last one
    if (index === newEntries.length - 1 && (key !== '' || value !== '')) {
      newEntries.push({ key: '', value: '' });
    }

    setEntries(newEntries);
    
    // Propagate to parent
    const newData: Record<string, string> = {};
    newEntries.forEach(e => {
      if (e.key.trim() !== '') {
        newData[e.key.trim()] = e.value;
      }
    });
    onChange(newData);
  };

  const handleDelete = (index: number) => {
    if (index === entries.length - 1) return; // Don't delete the last empty row
    const newEntries = entries.filter((_, i) => i !== index);
    setEntries(newEntries);
    
    const newData: Record<string, string> = {};
    newEntries.forEach(e => {
      if (e.key.trim() !== '') {
        newData[e.key.trim()] = e.value;
      }
    });
    onChange(newData);
  };

  return (
    <div className="flex flex-col">
      <div className="border-b border-[#444] bg-[#222]" style={gridStyle}>
        <div className="relative p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">
          Key
          <div 
            className="absolute top-0 right-0 w-1.5 h-full cursor-col-resize hover:bg-[#5BA76B] z-10 transition-colors" 
            onMouseDown={(e) => handleMouseDown(e, 0)}
          />
        </div>
        <div className="relative p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">
          Value
          <div 
            className="absolute top-0 right-0 w-1.5 h-full cursor-col-resize hover:bg-[#5BA76B] z-10 transition-colors" 
            onMouseDown={(e) => handleMouseDown(e, 1)}
          />
        </div>
        <div className="p-3 text-xs uppercase tracking-widest font-semibold">Description</div>
      </div>
      
      {entries.map((entry, i) => (
        <div key={i} className="border-b border-[#444] group bg-[#111111]" style={gridStyle}>
          <div className="border-r border-[#444]">
            <input 
              type="text" 
              className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" 
              value={entry.key}
              placeholder="Key..."
              onChange={e => handleChange(i, e.target.value, entry.value)}
            />
          </div>
          <div className="border-r border-[#444]">
            <input 
              type="text" 
              className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" 
              value={entry.value}
              placeholder="Value..."
              onChange={e => handleChange(i, entry.key, e.target.value)}
            />
          </div>
          <div className="relative flex items-center">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] text-[#888] focus:outline-none focus:bg-[#1A1A1A]" placeholder="Description..." />
            {i !== entries.length - 1 && (
              <button 
                className="absolute right-3 opacity-0 group-hover:opacity-100 text-[#E84E38] hover:text-[#ff6b57] transition-opacity"
                onClick={() => handleDelete(i)}
                title="Delete row"
              >
                ✕
              </button>
            )}
          </div>
        </div>
      ))}
      
      {/* Fill empty space if we have less than 3 rows total */}
      {[...Array(Math.max(3 - entries.length, 0))].map((_, i) => (
        <div key={`empty-fill-${i}`} className="border-b border-[#444] last:border-b-0 group bg-[#111111] opacity-50" style={gridStyle}>
          <div className="border-r border-[#444] p-3 text-sm h-[45px]"></div>
          <div className="border-r border-[#444] p-3 text-sm h-[45px]"></div>
          <div className="p-3 text-sm h-[45px]"></div>
        </div>
      ))}
    </div>
  );
}

function AuthEditor({ auth, setAuth }: { auth: models.Auth | undefined, setAuth: (a: models.Auth | undefined) => void }) {
  const isInherited = auth === undefined;
  const authType = auth ? auth.Type : 'none';

  const handleModeChange = (inherited: boolean) => {
    if (inherited) {
      setAuth(undefined);
    } else {
      setAuth(new models.Auth({ Type: 'none', Params: {} }));
    }
  };

  const handleTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const type = e.target.value;
    setAuth(new models.Auth({ Type: type, Params: {} }));
  };

  const updateParam = (key: string, value: string) => {
    if (!auth) return;
    setAuth(new models.Auth({
      Type: auth.Type,
      Params: { ...auth.Params, [key]: value }
    }));
  };

  return (
    <div className="p-6 text-sm font-['JetBrains_Mono'] text-[#F5F4F0] h-full overflow-auto custom-scrollbar">
      <div className="mb-6 flex items-center gap-6 text-xs uppercase tracking-widest font-semibold text-[#888]">
        <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
          <input 
            type="radio" 
            checked={isInherited} 
            onChange={() => handleModeChange(true)} 
            className="accent-[#F5F4F0]" 
          /> 
          Inherit from Workspace
        </label>
        <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
          <input 
            type="radio" 
            checked={!isInherited} 
            onChange={() => handleModeChange(false)} 
            className="accent-[#F5F4F0]" 
          /> 
          Custom Auth
        </label>
      </div>

      {!isInherited && (
        <div className="space-y-6 max-w-2xl">
          <div className="flex flex-col gap-2">
            <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Auth Type</label>
            <select 
              className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full"
              value={authType}
              onChange={handleTypeChange}
            >
              <option value="none">None</option>
              <option value="bearer">Bearer Token</option>
              <option value="basic">Basic Auth</option>
              <option value="apikey">API Key</option>
            </select>
          </div>

          {authType === 'bearer' && (
            <div className="flex flex-col gap-2">
              <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Token</label>
              <input 
                type="text" 
                className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full"
                placeholder="{{MY_TOKEN}}"
                value={auth?.Params?.token || ''}
                onChange={e => updateParam('token', e.target.value)}
              />
              <p className="text-xs text-[#888] mt-1">Supports variable interpolation like {'{{TOKEN}}'}</p>
            </div>
          )}

          {authType === 'basic' && (
            <>
              <div className="flex flex-col gap-2">
                <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Username</label>
                <input 
                  type="text" 
                  className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full"
                  placeholder="admin"
                  value={auth?.Params?.username || ''}
                  onChange={e => updateParam('username', e.target.value)}
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Password</label>
                <input 
                  type="password" 
                  className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full font-sans"
                  placeholder="{{DB_PASSWORD}}"
                  value={auth?.Params?.password || ''}
                  onChange={e => updateParam('password', e.target.value)}
                />
                <p className="text-xs text-[#888] mt-1">Use variables for passwords so they are not saved in plain text.</p>
              </div>
            </>
          )}

          {authType === 'apikey' && (
            <>
              <div className="flex flex-col gap-2">
                <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Header Name</label>
                <input 
                  type="text" 
                  className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full"
                  placeholder="X-API-Key"
                  value={auth?.Params?.header || ''}
                  onChange={e => updateParam('header', e.target.value)}
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs uppercase tracking-widest font-semibold text-[#888]">Key Value</label>
                <input 
                  type="text" 
                  className="p-3 bg-[#1A1A1A] border border-[#444] text-[#F5F4F0] focus:outline-none focus:border-[#5BA76B] w-full"
                  placeholder="{{API_KEY}}"
                  value={auth?.Params?.key || ''}
                  onChange={e => updateParam('key', e.target.value)}
                />
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
