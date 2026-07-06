// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React from 'react';
import beautify from 'js-beautify';
import { Check, Clock } from 'lucide-react';
import { models } from '../../../../wailsjs/go/models';

interface ResponseBoxProps {
  response: models.APIResponse | null;
}

export default function ResponseBox({ response }: ResponseBoxProps) {
  const renderBody = () => {
    if (!response) return 'Send a request to see the response.';
    if (!response.Body) return '<No Content>';

    if (response.ContentType === 'image') {
      const mime = (response.Headers && response.Headers['Content-Type'] && response.Headers['Content-Type'][0]) || 'image/png';
      return (
        <div className="flex items-center justify-center p-8">
          <img src={`data:${mime};base64,${response.Body}`} alt="Response output" className="max-w-full h-auto border border-[#444] shadow-lg" />
        </div>
      );
    }

    if (response.ContentType === 'pdf') {
      const mime = (response.Headers && response.Headers['Content-Type'] && response.Headers['Content-Type'][0]) || 'application/pdf';
      return (
        <div className="flex w-full h-[500px]">
          <iframe src={`data:${mime};base64,${response.Body}`} className="w-full h-full border-none" title="PDF Response" />
        </div>
      );
    }

    let displayBody = response.Body;
    if (response.ContentType === 'json' || (response.Headers && response.Headers['Content-Type']?.[0]?.includes('json'))) {
      try {
        displayBody = JSON.stringify(JSON.parse(displayBody), null, 2);
      } catch (e) {}
    } else if (response.ContentType === 'html' || (response.Headers && response.Headers['Content-Type']?.[0]?.includes('html'))) {
      try {
        displayBody = beautify.html(displayBody, { indent_size: 2 });
      } catch (e) {}
    } else if (response.ContentType === 'css' || (response.Headers && response.Headers['Content-Type']?.[0]?.includes('css'))) {
      try {
        displayBody = beautify.css(displayBody, { indent_size: 2 });
      } catch (e) {}
    } else if (response.ContentType === 'javascript' || (response.Headers && response.Headers['Content-Type']?.[0]?.includes('javascript'))) {
      try {
        displayBody = beautify.js(displayBody, { indent_size: 2 });
      } catch (e) {}
    }

    return (
      <pre className="whitespace-pre-wrap break-all">{displayBody}</pre>
    );
  };

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
                <span>{(response.Latency / 1000000000).toFixed(2)}s</span>
              </div>
            </div>
          )}
        </div>
      </div>
      
      <div className="flex-1 p-6 overflow-auto">
        <div className="border border-[#444] p-6 min-h-full font-['JetBrains_Mono'] text-sm text-[#F5F4F0] overflow-auto bg-[#1A1A1A] shadow-[4px_4px_0px_0px_rgba(68,68,68,1)]">
          {renderBody()}
        </div>
      </div>
    </div>
  );
}
