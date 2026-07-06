// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React from 'react';
import { FolderOpen, Home, FileText } from 'lucide-react';

interface SplashScreenProps {
  onSelectWorkspace: () => void;
  onUseDefaultWorkspace: () => void;
  onOpenSingleFile: () => void;
}

export default function SplashScreen({ onSelectWorkspace, onUseDefaultWorkspace, onOpenSingleFile }: SplashScreenProps) {
  return (
    <div className="flex h-screen w-full bg-[#111111] text-[#F5F4F0] flex-col items-center justify-center selection:bg-[#E84E38] selection:text-white" style={{ fontFamily: 'Inter, sans-serif' }}>
      
      <div className="text-center mb-16 animate-fade-in-up">
        <h1 className="font-['Playfair_Display'] font-bold text-7xl tracking-tight text-[#F5F4F0] mb-4">Anjal.</h1>
        <p className="text-[#888] text-sm uppercase tracking-[0.3em] font-semibold">The API Client Workspace</p>
      </div>

      <div className="flex flex-col sm:flex-row gap-6 max-w-4xl w-full px-6">
        <button 
          onClick={onSelectWorkspace}
          className="flex-1 group flex flex-col items-center gap-4 bg-[#1A1A1A] border border-[#444] p-10 hover:border-[#E84E38] hover:-translate-y-1 transition-all shadow-[8px_8px_0px_0px_rgba(68,68,68,1)] hover:shadow-[12px_12px_0px_0px_rgba(232,78,56,0.2)]"
        >
          <div className="w-16 h-16 rounded-full bg-[#222] flex items-center justify-center group-hover:bg-[#E84E38]/10 transition-colors">
            <FolderOpen className="w-8 h-8 text-[#888] group-hover:text-[#E84E38]" />
          </div>
          <div className="text-center">
            <h3 className="font-['Playfair_Display'] italic text-2xl font-semibold mb-2 text-[#F5F4F0]">Open Workspace</h3>
            <p className="text-xs text-[#888] font-['JetBrains_Mono'] leading-relaxed">
              Select an existing local folder to load or initialize as an Anjal workspace.
            </p>
          </div>
        </button>

        <button 
          onClick={onUseDefaultWorkspace}
          className="flex-1 group flex flex-col items-center gap-4 bg-[#1A1A1A] border border-[#444] p-10 hover:border-[#5BA76B] hover:-translate-y-1 transition-all shadow-[8px_8px_0px_0px_rgba(68,68,68,1)] hover:shadow-[12px_12px_0px_0px_rgba(91,167,107,0.2)]"
        >
          <div className="w-16 h-16 rounded-full bg-[#222] flex items-center justify-center group-hover:bg-[#5BA76B]/10 transition-colors">
            <Home className="w-8 h-8 text-[#888] group-hover:text-[#5BA76B]" />
          </div>
          <div className="text-center">
            <h3 className="font-['Playfair_Display'] italic text-2xl font-semibold mb-2 text-[#F5F4F0]">Use Default</h3>
            <p className="text-xs text-[#888] font-['JetBrains_Mono'] leading-relaxed">
              Load your global default workspace located in the home directory {navigator.userAgent.includes('Windows') ? '(%USERPROFILE%\\.anjal)' : '(~/.anjal)'}.
            </p>
          </div>
        </button>

        <button 
          onClick={onOpenSingleFile}
          className="flex-1 group flex flex-col items-center gap-4 bg-[#1A1A1A] border border-[#444] p-10 hover:border-[#3B82F6] hover:-translate-y-1 transition-all shadow-[8px_8px_0px_0px_rgba(68,68,68,1)] hover:shadow-[12px_12px_0px_0px_rgba(59,130,246,0.2)]"
        >
          <div className="w-16 h-16 rounded-full bg-[#222] flex items-center justify-center group-hover:bg-[#3B82F6]/10 transition-colors">
            <FileText className="w-8 h-8 text-[#888] group-hover:text-[#3B82F6]" />
          </div>
          <div className="text-center">
            <h3 className="font-['Playfair_Display'] italic text-2xl font-semibold mb-2 text-[#F5F4F0]">Open File Only</h3>
            <p className="text-xs text-[#888] font-['JetBrains_Mono'] leading-relaxed">
              Open a single .md collection file directly without loading a full workspace.
            </p>
          </div>
        </button>
      </div>

    </div>
  );
}
