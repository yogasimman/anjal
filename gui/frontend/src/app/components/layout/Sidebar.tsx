// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React, { useState } from 'react';
import { Search, FolderOpen, CornerDownRight, Plus, Trash2, Settings } from 'lucide-react';
import { models } from '../../../../wailsjs/go/models';

interface SidebarProps {
  isOpen: boolean;
  collections: models.Collection[];
  activeRequest: models.APIRequest | null;
  onOpenCollection: () => void;
  onSelectWorkspace: () => void;
  onSelectRequest: (req: models.APIRequest) => void;
  onAddRequest: (col: models.Collection) => void;
  onAddCollection: () => void;
  onDeleteRequest: (col: models.Collection, req: models.APIRequest) => void;
  onDeleteCollection: (col: models.Collection) => void;
  onRunAll?: (col: models.Collection) => void;
  isMultiRunning?: boolean;
  isSingleFileMode?: boolean;
  onOpenSettings: () => void;
  onCloseWorkspace: () => void;
}

export default function Sidebar({ 
  isOpen, collections, activeRequest, onOpenCollection, onSelectWorkspace, 
  onSelectRequest, onAddRequest, onAddCollection, onDeleteRequest, onDeleteCollection, onRunAll, isMultiRunning, isSingleFileMode, onOpenSettings, onCloseWorkspace
}: SidebarProps) {
  return (
    <aside className={`${isOpen ? 'w-72' : 'w-0 -translate-x-full'} overflow-hidden transition-all duration-500 flex-shrink-0 border-r border-[#444] flex flex-col relative z-20 bg-[#111111]`}>
      <div className="p-6 border-b border-[#444] flex items-center justify-between">
        <div className="flex flex-col">
          <span className="font-['Playfair_Display'] font-bold text-4xl tracking-tight text-[#F5F4F0]">Anjal.</span>
          <span className="text-[10px] uppercase tracking-[0.2em] text-[#888] mt-1">API Client Workspace</span>
        </div>
        <div className="flex">
          <button onClick={onSelectWorkspace} className="p-2 hover:bg-[#222] rounded transition-colors" title="Select Workspace Folder">
            <FolderOpen className="w-5 h-5 text-[#888] hover:text-[#5BA76B]" />
          </button>
          <button onClick={onCloseWorkspace} className="p-2 hover:bg-[#222] rounded transition-colors ml-1" title="Close Workspace">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-[#888] hover:text-[#E84E38]"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar">
        <div className="px-4 py-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="font-['Playfair_Display'] italic text-xl text-[#F5F4F0]">Collections</h2>
            {!isSingleFileMode && (
              <button onClick={onAddCollection} className="p-1 hover:bg-[#222] rounded text-[#888] hover:text-[#5BA76B] transition-colors" title="Add Collection">
                <Plus className="w-4 h-4" />
              </button>
            )}
          </div>
          <div className="space-y-6">
            {collections.map(col => (
              <CollectionItem 
                key={col.FilePath} 
                title={col.Name} 
                defaultOpen={true}
                onAdd={() => onAddRequest(col)}
                onDelete={() => onDeleteCollection(col)}
                onRunAll={() => onRunAll && onRunAll(col)}
                isMultiRunning={isMultiRunning}
                isSingleFileMode={isSingleFileMode}
              >
                {col.Requests && col.Requests.map(req => (
                  <div key={req.ID}>
                    <RequestItem 
                      method={req.Method} 
                      name={req.Title} 
                      active={activeRequest?.ID === req.ID} 
                      onClick={() => onSelectRequest(req)}
                      onDelete={(e) => { e.stopPropagation(); onDeleteRequest(col, req); }}
                    />
                  </div>
                ))}
              </CollectionItem>
            ))}
            {collections.length === 0 && (
              <div className="text-sm text-[#666] italic px-2">No collections found. Open a workspace.</div>
            )}
          </div>
        </div>
      </div>

      <div className="p-4 border-t border-[#444] text-xs uppercase tracking-widest flex justify-between items-center text-[#F5F4F0] font-semibold">
        <button onClick={onOpenSettings} className="hover:text-[#E84E38] transition-colors flex items-center gap-2">
          <Settings className="w-4 h-4" /> Environment
        </button>
        <button className="hover:text-[#E84E38] transition-colors">History</button>
      </div>
    </aside>
  );
}

function CollectionItem({ title, children, defaultOpen = false, onAdd, onDelete, onRunAll, isMultiRunning, isSingleFileMode }: { title: string, children: React.ReactNode, defaultOpen?: boolean, onAdd: () => void, onDelete: () => void, onRunAll?: () => void, isMultiRunning?: boolean, isSingleFileMode?: boolean }) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  
  return (
    <div className="mb-4">
      <div className="flex items-center justify-between group">
        <div 
          className="flex items-center gap-3 py-1 cursor-pointer flex-1"
          onClick={() => setIsOpen(!isOpen)}
        >
          <div className="w-4 h-px bg-[#F5F4F0] transition-all group-hover:w-6"></div>
          <span className="font-semibold text-lg">{title}</span>
        </div>
        <div className="opacity-0 group-hover:opacity-100 flex items-center transition-opacity">
          {onRunAll && (
            <button onClick={(e) => { e.stopPropagation(); onRunAll(); }} disabled={isMultiRunning} className={`p-1 ${isMultiRunning ? 'text-[#5BA76B]' : 'text-[#888] hover:text-[#5BA76B]'}`} title="Run All Requests">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
            </button>
          )}
          <button onClick={(e) => { e.stopPropagation(); onAdd(); }} className="p-1 text-[#888] hover:text-[#5BA76B]" title="Add Request">
            <Plus className="w-4 h-4" />
          </button>
          {!isSingleFileMode && (
            <button onClick={(e) => { e.stopPropagation(); onDelete(); }} className="p-1 text-[#888] hover:text-[#E84E38]" title="Delete Collection">
              <Trash2 className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>
      {isOpen && (
        <div className="ml-4 mt-3 space-y-2 border-l border-[#444] pl-4 pr-2">
          {children}
        </div>
      )}
    </div>
  );
}

function RequestItem({ method, name, active = false, onClick, onDelete }: { method: string, name: string, active?: boolean, onClick: () => void, onDelete: (e: React.MouseEvent) => void }) {
  const methodColor = 
    method === 'GET' ? 'text-[#5BA76B]' : 
    method === 'POST' ? 'text-[#E09A38]' : 
    method === 'PUT' ? 'text-[#3B82F6]' : 
    method === 'PATCH' ? 'text-[#A855F7]' : 
    method === 'DELETE' ? 'text-[#E84E38]' : 'text-[#888]';

  return (
    <div 
      className={`group flex items-center justify-between py-1 cursor-pointer transition-all ${active ? 'translate-x-2' : 'hover:translate-x-1'}`}
      onClick={onClick}
    >
      <div className="flex items-center gap-3">
        {active && <CornerDownRight className="w-3 h-3 text-[#E84E38] absolute -left-4" />}
        <span className={`text-[10px] font-['JetBrains_Mono'] font-bold w-12 ${methodColor}`}>{method}</span>
        <span className={`text-sm ${active ? 'font-bold text-[#F5F4F0]' : 'text-[#aaa]'}`}>{name}</span>
      </div>
      <button onClick={onDelete} className="opacity-0 group-hover:opacity-100 p-1 text-[#888] hover:text-[#E84E38] transition-opacity" title="Delete Request">
        <Trash2 className="w-3 h-3" />
      </button>
    </div>
  );
}
