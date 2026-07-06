// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React, { useState } from 'react';
import { 
  Search, Check, Clock, Database, 
  Menu, X, CornerDownRight, ArrowRight
} from 'lucide-react';

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

export default function App() {
  const [method, setMethod] = useState('GET');
  const [url, setUrl] = useState('https://api.anjal.dev/v1/users');
  const [activeTab, setActiveTab] = useState('Params');
  const [sidebarOpen, setSidebarOpen] = useState(true);

  return (
    <div className="flex h-screen w-full bg-[#111111] text-[#F5F4F0] selection:bg-[#E84E38] selection:text-white overflow-hidden" style={{ fontFamily: 'Inter, sans-serif' }}>
      
      {/* Sidebar - Brutalist/Editorial in Dark Mode */}
      <aside className={`${sidebarOpen ? 'w-72' : 'w-0 -translate-x-full'} transition-all duration-500 flex-shrink-0 border-r border-[#444] flex flex-col relative z-20 bg-[#111111]`}>
        <div className="p-6 border-b border-[#444] flex items-center justify-between">
          <div className="flex flex-col">
            <span className="font-['Playfair_Display'] font-bold text-4xl tracking-tight text-[#F5F4F0]">Anjal.</span>
            <span className="text-[10px] uppercase tracking-[0.2em] text-[#888] mt-1">API Client Workspace</span>
          </div>
        </div>

        <div className="p-4 border-b border-[#444]">
          <div className="relative group">
            <Search className="w-4 h-4 absolute left-0 top-1/2 -translate-y-1/2 text-[#F5F4F0]" />
            <input 
              type="text" 
              placeholder="Search collections..." 
              className="w-full bg-transparent text-sm text-[#F5F4F0] pl-6 pr-0 py-2 border-b border-transparent focus:outline-none focus:border-[#F5F4F0] transition-all placeholder:text-[#666] rounded-none"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar">
          <div className="px-4 py-6">
            <h2 className="font-['Playfair_Display'] italic text-xl text-[#F5F4F0] mb-6">Index</h2>
            <div className="space-y-6">
              {/* Collection Titles are Sans-Serif and have no Roman numerals per request */}
              <CollectionItem title="Authentication" defaultOpen={true}>
                <RequestItem method="POST" name="Login User" active />
                <RequestItem method="POST" name="Register Account" />
                <RequestItem method="POST" name="Refresh Token" />
              </CollectionItem>
              <CollectionItem title="Payments Service">
                <RequestItem method="GET" name="List Transactions" />
                <RequestItem method="POST" name="Create Charge" />
              </CollectionItem>
              <CollectionItem title="User Management">
                <RequestItem method="GET" name="Fetch Profile" />
                <RequestItem method="PATCH" name="Update Profile" />
                <RequestItem method="DELETE" name="Delete Account" />
              </CollectionItem>
            </div>
          </div>
        </div>

        <div className="p-4 border-t border-[#444] text-xs uppercase tracking-widest flex justify-between items-center text-[#F5F4F0] font-semibold">
          <button className="hover:text-[#E84E38] transition-colors">Settings</button>
          <button className="hover:text-[#E84E38] transition-colors">History</button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 bg-[#111111]">
        
        {/* Header / URL Bar - Brutalist */}
        <header className="border-b border-[#444] flex flex-col sm:flex-row sm:items-center bg-[#111111]">
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
              className={`appearance-none bg-transparent px-6 py-4 text-xl font-['Playfair_Display'] font-bold focus:outline-none cursor-pointer rounded-none
                ${method === 'GET' ? 'text-[#F5F4F0]' : 'text-[#E84E38]'}
              `}
            >
              {METHODS.map(m => <option key={m} value={m} className="bg-[#111111] text-[#F5F4F0] font-sans text-base">{m}</option>)}
            </select>
          </div>
          
          <div className="flex-1 flex items-center border-t sm:border-t-0 border-[#444]">
            <input 
              type="text" 
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="flex-1 bg-transparent border-none text-[#F5F4F0] text-lg px-6 py-4 focus:outline-none font-['JetBrains_Mono'] placeholder:text-[#666]"
              placeholder="https://"
            />
          </div>

          <div className="flex border-t sm:border-t-0 sm:border-l border-[#444]">
            <button className="bg-[#F5F4F0] hover:bg-[#DCDAD2] text-[#111111] font-bold px-8 py-4 text-sm tracking-widest uppercase flex items-center gap-3 transition-colors">
              Execute <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </header>

        {/* Split View Container */}
        <div className="flex-1 flex flex-col lg:flex-row overflow-hidden">
          
          {/* Request Section */}
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
              {activeTab === 'Params' && <KeyValueEditor />}
              {activeTab === 'Headers' && <KeyValueEditor headers />}
              {activeTab === 'Body' && (
                <div className="h-full flex flex-col p-6">
                  <div className="flex items-center gap-6 mb-6 text-xs uppercase tracking-widest font-semibold text-[#888]">
                    <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
                      <input type="radio" name="bodyType" className="accent-[#F5F4F0]" /> None
                    </label>
                    <label className="flex items-center gap-2 cursor-pointer text-[#F5F4F0]">
                      <input type="radio" name="bodyType" defaultChecked className="accent-[#F5F4F0]" /> Raw / JSON
                    </label>
                    <label className="flex items-center gap-2 cursor-pointer hover:text-[#F5F4F0] transition-colors">
                      <input type="radio" name="bodyType" className="accent-[#F5F4F0]" /> Form Data
                    </label>
                  </div>
                  <div className="flex-1 border border-[#444] p-6 font-['JetBrains_Mono'] text-sm overflow-auto text-[#F5F4F0] bg-[#1A1A1A] relative shadow-[4px_4px_0px_0px_rgba(68,68,68,1)] focus-within:shadow-[6px_6px_0px_0px_rgba(232,78,56,1)] transition-shadow">
                    <pre className="outline-none min-h-full" contentEditable suppressContentEditableWarning>
<span className="text-[#888]">{`{`}</span>{'\n'}
  <span className="text-[#F5F4F0] font-bold">"email"</span>: <span className="text-[#E84E38]">"user@anjal.dev"</span>,{'\n'}
  <span className="text-[#F5F4F0] font-bold">"password"</span>: <span className="text-[#E84E38]">"securepassword123"</span>,{'\n'}
  <span className="text-[#F5F4F0] font-bold">"preferences"</span>: <span className="text-[#888]">{`{`}</span>{'\n'}
    <span className="text-[#F5F4F0] font-bold">"theme"</span>: <span className="text-[#E84E38]">"dark_brutalist"</span>,{'\n'}
    <span className="text-[#F5F4F0] font-bold">"typography"</span>: <span className="text-[#E84E38]">"serif"</span>{'\n'}
  <span className="text-[#888]">{`}`}</span>{'\n'}
<span className="text-[#888]">{`}`}</span>
                    </pre>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Response Section */}
          <div className="flex-1 flex flex-col min-h-0 bg-[#111111]">
            <div className="border-b border-[#444] flex items-center justify-between px-6 py-3">
              <div className="flex items-center gap-8">
                <span className="font-['Playfair_Display'] italic text-xl font-semibold text-[#F5F4F0]">Response</span>
                <div className="flex items-center gap-6 text-xs font-['JetBrains_Mono']">
                  <div className="flex items-center gap-2 text-[#5BA76B]">
                    <Check className="w-4 h-4" />
                    <span>200 OK</span>
                  </div>
                  <div className="flex items-center gap-2 text-[#888]">
                    <Clock className="w-4 h-4" />
                    <span>142ms</span>
                  </div>
                  <div className="flex items-center gap-2 text-[#888]">
                    <Database className="w-4 h-4" />
                    <span>2.4kb</span>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="flex-1 p-6 overflow-auto">
              <div className="border border-[#444] p-6 min-h-full font-['JetBrains_Mono'] text-sm text-[#F5F4F0] overflow-auto bg-[#1A1A1A] shadow-[4px_4px_0px_0px_rgba(68,68,68,1)]">
                <pre>
                  <span className="text-[#888]">{`{`}</span>{'\n'}
                  <span className="text-[#F5F4F0] font-bold">  "status":</span> <span className="text-[#5BA76B]">"success"</span>,<span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#F5F4F0] font-bold">  "data":</span> <span className="text-[#888]">{`{`}</span>{'\n'}
                  <span className="text-[#F5F4F0] font-bold">    "user":</span> <span className="text-[#888]">{`{`}</span>{'\n'}
                  <span className="text-[#F5F4F0] font-bold">      "id":</span> <span className="text-[#E84E38]">"usr_98a7sd98f7"</span>,<span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#F5F4F0] font-bold">      "email":</span> <span className="text-[#5BA76B]">"user@anjal.dev"</span>,<span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#F5F4F0] font-bold">      "role":</span> <span className="text-[#5BA76B]">"admin"</span>,<span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#F5F4F0] font-bold">      "created_at":</span> <span className="text-[#5BA76B]">"2023-10-24T10:00:00Z"</span><span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#888]">    {`}`}</span>,<span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#F5F4F0] font-bold">    "token":</span> <span className="text-[#5BA76B]">"eyJhbGciOiJIUzI1NiIsInR..."</span><span className="text-[#888]">{'\n'}</span>
                  <span className="text-[#888]">  {`}`}</span>{'\n'}
                  <span className="text-[#888]">{`}`}</span>
                </pre>
              </div>
            </div>
          </div>

        </div>
        
        {/* Footer StatusBar */}
        <footer className="border-t border-[#444] bg-[#222] flex items-center px-6 py-2 justify-between text-xs font-['JetBrains_Mono'] text-[#F5F4F0]">
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-[#5BA76B]"></div> System Online</span>
            <span className="text-[#888]">Env: Production</span>
          </div>
          <div className="flex items-center gap-6">
            <span className="hover:underline cursor-pointer">Shortcuts</span>
            <span className="hover:underline cursor-pointer">Docs</span>
          </div>
        </footer>
      </main>

    </div>
  );
}

function CollectionItem({ title, children, defaultOpen = false }: { title: string, children: React.ReactNode, defaultOpen?: boolean }) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  
  return (
    <div className="mb-4">
      <div 
        className="flex items-center gap-3 py-1 cursor-pointer group"
        onClick={() => setIsOpen(!isOpen)}
      >
        <div className="w-4 h-px bg-[#F5F4F0] transition-all group-hover:w-6"></div>
        {/* Removed Serif Font from here, using standard Inter sans-serif */}
        <span className="font-semibold text-lg">{title}</span>
      </div>
      {isOpen && (
        <div className="ml-4 mt-3 space-y-2 border-l border-[#444] pl-4">
          {children}
        </div>
      )}
    </div>
  );
}

function RequestItem({ method, name, active = false }: { method: string, name: string, active?: boolean }) {
  const methodColor = 
    method === 'GET' ? 'text-[#F5F4F0]' : 
    method === 'POST' ? 'text-[#E84E38]' : 
    method === 'PUT' ? 'text-[#E09A38]' : 
    method === 'DELETE' ? 'text-[#C55A5A]' : 'text-[#888]';

  return (
    <div className={`flex items-center gap-3 py-1 cursor-pointer transition-all
      ${active ? 'translate-x-2' : 'hover:translate-x-1'}
    `}>
      {active && <CornerDownRight className="w-3 h-3 text-[#E84E38] absolute -left-4" />}
      <span className={`text-[10px] font-['JetBrains_Mono'] font-bold w-12 ${methodColor}`}>{method}</span>
      <span className={`text-sm ${active ? 'font-bold text-[#F5F4F0]' : 'text-[#aaa]'}`}>{name}</span>
    </div>
  );
}

function KeyValueEditor({ headers = false }: { headers?: boolean }) {
  const rows = headers ? 4 : 3;
  return (
    <div className="flex flex-col">
      <div className="grid grid-cols-12 border-b border-[#444] bg-[#222]">
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">Key</div>
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold border-r border-[#444]">Value</div>
        <div className="col-span-4 p-3 text-xs uppercase tracking-widest font-semibold">Description</div>
      </div>
      {[...Array(rows)].map((_, i) => (
        <div key={i} className="grid grid-cols-12 border-b border-[#444] last:border-b-0 group bg-[#111111]">
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" placeholder="Key..." 
              defaultValue={headers && i === 0 ? 'Authorization' : headers && i === 1 ? 'Content-Type' : ''} 
            />
          </div>
          <div className="col-span-4 border-r border-[#444]">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] focus:outline-none focus:bg-[#1A1A1A] text-[#F5F4F0]" placeholder="Value..." 
              defaultValue={headers && i === 0 ? 'Bearer eyJhb...' : headers && i === 1 ? 'application/json' : ''}
            />
          </div>
          <div className="col-span-4 relative flex items-center">
            <input type="text" className="w-full bg-transparent p-3 text-sm font-['JetBrains_Mono'] text-[#888] focus:outline-none focus:bg-[#1A1A1A]" placeholder="Description..." />
            <button className="absolute right-3 opacity-0 group-hover:opacity-100 text-[#E84E38] transition-all">
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
