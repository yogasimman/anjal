import React, { useState } from 'react';
import { X, Plus, Trash2 } from 'lucide-react';

interface EnvironmentModalProps {
  isOpen: boolean;
  onClose: () => void;
  env: Record<string, string>;
  setEnv: (e: Record<string, string>) => void;
}

export default function EnvironmentModal({ isOpen, onClose, env, setEnv }: EnvironmentModalProps) {
  const [entries, setEntries] = useState(Object.entries(env));

  if (!isOpen) return null;

  const handleAdd = () => {
    setEntries([...entries, ['', '']]);
  };

  const handleUpdate = (index: number, key: string, value: string) => {
    const newEntries = [...entries];
    newEntries[index] = [key, value];
    setEntries(newEntries);
  };

  const handleRemove = (index: number) => {
    const newEntries = entries.filter((_, i) => i !== index);
    setEntries(newEntries);
  };

  const handleSave = () => {
    const newEnv: Record<string, string> = {};
    for (const [k, v] of entries) {
      if (k.trim() !== '') {
        newEnv[k.trim()] = v;
      }
    }
    setEnv(newEnv);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-[#111111] border border-[#444] shadow-[8px_8px_0px_0px_rgba(68,68,68,1)] w-full max-w-2xl text-[#F5F4F0] flex flex-col">
        <div className="flex justify-between items-center border-b border-[#444] p-4 bg-[#222]">
          <h2 className="font-['Playfair_Display'] italic text-2xl font-bold">Workspace Environment</h2>
          <button onClick={onClose} className="hover:text-[#E84E38] transition-colors"><X className="w-6 h-6" /></button>
        </div>
        
        <div className="p-6 flex-1 overflow-auto max-h-[60vh] space-y-4">
          <p className="text-sm text-[#888] font-['JetBrains_Mono'] mb-4">
            Variables are stored locally in the app. Use {"{{"}VARIABLE{"}}"} syntax in your requests.
          </p>
          
          {entries.map(([k, v], index) => (
            <div key={index} className="flex gap-4 items-center">
              <input 
                type="text" 
                placeholder="Key" 
                value={k}
                onChange={(e) => handleUpdate(index, e.target.value, v)}
                className="flex-1 bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] border border-[#444] text-[#F5F4F0]" 
              />
              <input 
                type="text" 
                placeholder="Value" 
                value={v}
                onChange={(e) => handleUpdate(index, k, e.target.value)}
                className="flex-1 bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] border border-[#444] text-[#F5F4F0]" 
              />
              <button 
                onClick={() => handleRemove(index)}
                className="p-3 border border-[#444] hover:bg-[#E84E38] hover:border-[#E84E38] transition-colors"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}

          <button 
            onClick={handleAdd}
            className="flex items-center gap-2 text-sm font-['JetBrains_Mono'] uppercase tracking-widest hover:text-[#5BA76B] transition-colors"
          >
            <Plus className="w-4 h-4" /> Add Variable
          </button>
        </div>

        <div className="border-t border-[#444] p-4 flex justify-end gap-4 bg-[#222]">
          <button onClick={onClose} className="px-6 py-2 text-sm font-bold uppercase tracking-widest border border-[#444] hover:bg-[#444] transition-colors">
            Cancel
          </button>
          <button onClick={handleSave} className="bg-[#F5F4F0] text-[#111111] px-6 py-2 text-sm font-bold uppercase tracking-widest hover:bg-[#DCDAD2] transition-colors">
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}
