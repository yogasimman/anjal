import React from 'react';
import { Check, Clock } from 'lucide-react';
import { models } from '../../../../wailsjs/go/models';

interface ResponseBoxProps {
  response: models.APIResponse | null;
}

export default function ResponseBox({ response }: ResponseBoxProps) {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#111111]">
      <div className="border-b border-[#444] flex items-center justify-between px-6 py-3">
        <div className="flex items-center gap-8">
          <span className="font-['Playfair_Display'] italic text-xl font-semibold text-[#F5F4F0]">Response</span>
          {response && (
            <div className="flex items-center gap-6 text-xs font-['JetBrains_Mono']">
              <div className={`flex items-center gap-2 ${response.StatusCode < 400 ? 'text-[#5BA76B]' : 'text-[#E84E38]'}`}>
                <Check className="w-4 h-4" />
                <span>{response.Status || response.StatusCode}</span>
              </div>
              <div className="flex items-center gap-2 text-[#888]">
                <Clock className="w-4 h-4" />
                <span>{response.Latency}ms</span>
              </div>
            </div>
          )}
        </div>
      </div>
      
      <div className="flex-1 p-6 overflow-auto">
        <div className="border border-[#444] p-6 min-h-full font-['JetBrains_Mono'] text-sm text-[#F5F4F0] overflow-auto bg-[#1A1A1A] shadow-[4px_4px_0px_0px_rgba(68,68,68,1)]">
          <pre className="whitespace-pre-wrap">{response?.Body || 'Send a request to see the response.'}</pre>
        </div>
      </div>
    </div>
  );
}
