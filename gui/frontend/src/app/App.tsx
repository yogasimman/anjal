// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import React, { useState, useEffect } from 'react';
import { 
  GetCollections, 
  GetCollectionsFrom, 
  PromptOpenWorkspace,
  PromptOpenFile,
  ExecuteRequest,
  CreateRequest,
  UpdateRequest,
  DeleteRequest,
  DeleteCollection,
  LoadEnvForCollection,
  SaveEnvForCollection,
  DeleteEnvForCollection,
  PromptSaveFilePath,
  WriteFile,
  CreateCollection
} from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';

import Sidebar from './components/layout/Sidebar';
import TopUrlBar from './components/layout/TopUrlBar';
import RequestBox from './components/layout/RequestBox';
import ResponseBox from './components/layout/ResponseBox';
import EnvironmentModal from './components/layout/EnvironmentModal';
import SplashScreen from './components/layout/SplashScreen';

export default function App() {
  const [method, setMethod] = useState('GET');
  const [url, setUrl] = useState('https://api.anjal.dev/v1/users');
  const [title, setTitle] = useState('New Request');
  const [body, setBody] = useState('');
  const [headers, setHeaders] = useState<Record<string, string>>({});
  const [params, setParams] = useState<Record<string, string>>({});
  const [auth, setAuth] = useState<models.Auth | undefined>(undefined);
  const [drafts, setDrafts] = useState<Record<string, Partial<models.APIRequest>>>({});
  const [responseCache, setResponseCache] = useState<Record<string, models.APIResponse>>({});
  const [sidebarOpen, setSidebarOpen] = useState(true);

  // App State
  const [workspacePath, setWorkspacePath] = useState<string>('');
  const [isWorkspaceSelected, setIsWorkspaceSelected] = useState<boolean>(false);
  const [isSingleFileMode, setIsSingleFileMode] = useState<boolean>(false);
  const [collections, setCollections] = useState<models.Collection[]>([]);
  const [activeRequest, setActiveRequest] = useState<models.APIRequest | null>(null);
  const [activeCollection, setActiveCollection] = useState<models.Collection | null>(null);
  const [response, setResponse] = useState<models.APIResponse | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [multiRunResults, setMultiRunResults] = useState<{req: models.APIRequest, res: models.APIResponse}[] | null>(null);
  const [isMultiRunning, setIsMultiRunning] = useState(false);

  // Environment & Local Config
  const [envModalOpen, setEnvModalOpen] = useState(false);
  const [envConfig, setEnvConfig] = useState<Record<string, string>>({});

  useEffect(() => {
    const fetchEnv = async () => {
      try {
        const colName = activeCollection ? activeCollection.Name : "";
        const envs = await LoadEnvForCollection(colName);
        setEnvConfig(envs || {});
      } catch (e) {
        console.error("Failed to load env", e);
      }
    };
    fetchEnv();
  }, [activeCollection?.Name]);

  const handleSaveEnv = async (newEnv: Record<string, string>) => {
    const colName = activeCollection ? activeCollection.Name : "";
    
    // We do a simple diff to save/delete to Wails backend
    for (const [k, v] of Object.entries(newEnv)) {
      if (envConfig[k] !== v) {
        await SaveEnvForCollection(colName, k, v);
      }
    }
    for (const k of Object.keys(envConfig)) {
      if (!(k in newEnv)) {
        await DeleteEnvForCollection(colName, k);
      }
    }
    
    setEnvConfig(newEnv);
  };

  // Splash screen is now always shown on load by removing auto-load

  useEffect(() => {
    if (!isWorkspaceSelected || isSingleFileMode) return;
    if (workspacePath) {
      loadCollectionsFrom(workspacePath);
    } else {
      loadCollections();
    }
  }, [workspacePath, isWorkspaceSelected, isSingleFileMode]);

  const loadCollections = async () => {
    try {
      const cols = await GetCollections();
      if (cols) {
        setCollections(cols);
      } else {
        setCollections([]);
      }
    } catch (e) {
      console.error("Failed to load collections", e);
    }
  };

  const loadCollectionsFrom = async (dir: string) => {
    try {
      const cols = await GetCollectionsFrom(dir);
      if (cols) {
        setCollections(cols);
      } else {
        setCollections([]);
      }
    } catch (e) {
      console.error("Failed to load collections from", dir, e);
    }
  };

  const reloadCollections = async (targetFilePath?: string) => {
    if (isSingleFileMode) {
      const path = targetFilePath || activeCollection?.FilePath;
      if (!path) return;
      const dir = path.substring(0, Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'))) || '.';
      try {
        const cols = await GetCollectionsFrom(dir);
        if (cols) {
          const updatedCol = cols.find(c => c.FilePath === path);
          if (updatedCol) {
            setCollections([updatedCol]);
          }
        }
      } catch (e) {
        console.error("Failed to reload single file", e);
      }
      return;
    }
    if (workspacePath) loadCollectionsFrom(workspacePath);
    else loadCollections();
  };

  const handleSelectWorkspace = async () => {
    try {
      const dir = await PromptOpenWorkspace();
      if (dir) {
        setWorkspacePath(dir);
        localStorage.setItem('anjal-workspace', dir);
        setIsSingleFileMode(false);
        setIsWorkspaceSelected(true);
      }
    } catch (e) {
      console.error("Failed to open workspace", e);
    }
  };

  const handleOpenSingleFile = async () => {
    try {
      const col = await PromptOpenFile();
      if (col) {
        setWorkspacePath('');
        setCollections([col]);
        setIsSingleFileMode(true);
        setIsWorkspaceSelected(true);
      }
    } catch (e) {
      console.error("Failed to open single file", e);
      alert("Failed to load file: " + String(e));
    }
  };

  const handleCloseWorkspace = () => {
    localStorage.removeItem('anjal-workspace');
    setWorkspacePath('');
    setIsWorkspaceSelected(false);
    setIsSingleFileMode(false);
    setCollections([]);
    setActiveRequest(null);
    setActiveCollection(null);
    setResponse(null);
    setMultiRunResults(null);
  };

  const handleSelectRequest = (req: models.APIRequest) => {
    // Save current active request to drafts before switching
    if (activeRequest) {
      setDrafts(prev => ({
        ...prev,
        [activeRequest.ID]: {
          Method: method,
          URL: url,
          Title: title,
          Body: body,
          Headers: headers,
          QueryParams: params,
          Auth: auth
        }
      }));
    }

    setActiveRequest(req);
    // Find parent collection for saving updates later
    const col = collections.find(c => c.Requests?.some(r => r.ID === req.ID));
    if (col) setActiveCollection(col);

    setDrafts(prev => {
      const draft = prev[req.ID];
      setMethod(draft ? draft.Method || 'GET' : (req.Method || 'GET'));
      setUrl(draft ? draft.URL || '' : (req.URL || ''));
      setTitle(draft ? draft.Title || '' : (req.Title || 'New Request'));
      setBody(draft ? draft.Body || '' : (req.Body || ''));
      setHeaders(draft ? draft.Headers || {} : (req.Headers || {}));
      setParams(draft ? draft.QueryParams || {} : (req.QueryParams || {}));
      setAuth(draft ? draft.Auth : req.Auth);
      return prev;
    });
    setResponse(responseCache[req.ID] || null);
  };

  // CRUD Operations
  const handleAddRequest = async (col: models.Collection) => {
    const newReq = new models.APIRequest({
      ID: `req-${Date.now().toString(16)}`,
      Title: "New Request",
      Method: "GET",
      URL: "https://",
    });
    try {
      await CreateRequest(col.FilePath, newReq);
      // Reload collections
      reloadCollections(col.FilePath);
    } catch (e) {
      console.error("Failed to create request", e);
    }
  };

  const handleAddCollection = async () => {
    const name = prompt("Enter a name for the new collection:");
    if (!name || name.trim() === "") return;
    try {
      await CreateCollection(workspacePath || ".", name.trim());
      reloadCollections();
    } catch (e) {
      console.error("Failed to create collection", e);
      alert(String(e));
    }
  };

  const handleDeleteRequest = async (col: models.Collection, req: models.APIRequest) => {
    if (!confirm(`Are you sure you want to delete the request: ${req.Title}?`)) return;
    try {
      await DeleteRequest(col.FilePath, req.ID);
      
      // Clear draft to prevent memory leak
      setDrafts(prev => {
        const newDrafts = { ...prev };
        delete newDrafts[req.ID];
        return newDrafts;
      });

      if (activeRequest?.ID === req.ID) {
        setActiveRequest(null);
        setResponse(null);
      }
      reloadCollections(col.FilePath);
    } catch (e) {
      console.error("Failed to delete request", e);
    }
  };

  const handleDeleteCollection = async (col: models.Collection) => {
    if (!confirm(`Are you sure you want to delete the collection: ${col.Name}?`)) return;
    try {
      await DeleteCollection(col.FilePath);
      if (activeCollection?.FilePath === col.FilePath) {
        setActiveRequest(null);
        setActiveCollection(null);
        setResponse(null);
      }
      reloadCollections();
    } catch (e) {
      console.error("Failed to delete collection", e);
    }
  };

  // Execution & Variable Interpolation
  const interpolateVars = (str: string) => {
    if (!str) return str;
    let result = str;
    for (const [key, value] of Object.entries(envConfig)) {
      result = result.replace(new RegExp(`\\{\\{\\s*${key}\\s*\\}\\}`, 'g'), value);
    }
    return result;
  };

  const handleExecute = async () => {
    if (!activeRequest) return;
    setIsExecuting(true);
    setResponse(null);
    try {
      // 1. Save UI state back to the markdown via UpdateRequest
      const reqToExecute = new models.APIRequest({
        ...activeRequest,
        Title: title,
        Method: method,
        URL: url,
        Body: body,
        Headers: headers,
        QueryParams: params,
        Auth: auth,
      });

      // Clear the draft since it's now saved to the backend
      setDrafts(prev => {
        const newDrafts = { ...prev };
        delete newDrafts[activeRequest.ID];
        return newDrafts;
      });

      if (activeCollection) {
        await UpdateRequest(activeCollection.FilePath, activeRequest.ID, reqToExecute);
      }

      // 2. Resolve authentication if workspace configured it AND request doesn't override
      let authConfig = reqToExecute.Auth;
      if (!authConfig) {
        const authType = envConfig['WORKSPACE_AUTH_TYPE'];
        if (authType && authType !== 'none') {
          const params: Record<string, string> = {};
          for (const [k, v] of Object.entries(envConfig)) {
            if (k.startsWith('WORKSPACE_AUTH_') && k !== 'WORKSPACE_AUTH_TYPE') {
              const paramKey = k.replace('WORKSPACE_AUTH_', '').toLowerCase();
              params[paramKey] = v;
            }
          }
          authConfig = new models.Auth({ Type: authType, Params: params });
        }
      }

      // 3. Interpolate variables for execution (don't save interpolated back to MD)
      const executionReq = new models.APIRequest({
        ...reqToExecute,
        URL: interpolateVars(reqToExecute.URL),
        Body: interpolateVars(reqToExecute.Body || ''),
        Auth: authConfig,
      });

      // Interpolate auth params if they exist
      if (executionReq.Auth?.Params) {
        for (const [k, v] of Object.entries(executionReq.Auth.Params)) {
          executionReq.Auth.Params[k] = interpolateVars(v as string);
        }
      }

      const res = await ExecuteRequest(executionReq);
      setResponse(res);
      setResponseCache(prev => ({ ...prev, [activeRequest.ID]: res }));
      setMultiRunResults(null); // Clear multi-run results if single run is triggered
      
      // Reload to reflect saved changes in sidebar
      reloadCollections();

    } catch (e) {
      console.error(e);
      const errRes = new models.APIResponse({
        StatusCode: 0,
        Status: "Error",
        Body: String(e),
        Latency: 0,
        Headers: {},
        ContentType: "raw",
      });
      setResponse(errRes);
      setResponseCache(prev => ({ ...prev, [activeRequest.ID]: errRes }));
    } finally {
      setIsExecuting(false);
    }
  };

  const handleExecuteAll = async (col: models.Collection) => {
    if (!col.Requests || col.Requests.length === 0) return;
    setIsMultiRunning(true);
    const results: {req: models.APIRequest, res: models.APIResponse}[] = [];
    
    try {
      for (const req of col.Requests) {
        let authConfig = req.Auth;
        if (!authConfig) {
          const authType = envConfig['WORKSPACE_AUTH_TYPE'];
          if (authType && authType !== 'none') {
            const params: Record<string, string> = {};
            for (const [k, v] of Object.entries(envConfig)) {
              if (k.startsWith('WORKSPACE_AUTH_') && k !== 'WORKSPACE_AUTH_TYPE') {
                const paramKey = k.replace('WORKSPACE_AUTH_', '').toLowerCase();
                params[paramKey] = v;
              }
            }
            authConfig = new models.Auth({ Type: authType, Params: params });
          }
        }

        const executionReq = new models.APIRequest({
          ...req,
          URL: interpolateVars(req.URL),
          Body: interpolateVars(req.Body || ''),
          Auth: authConfig,
        });

        if (executionReq.Auth?.Params) {
          for (const [k, v] of Object.entries(executionReq.Auth.Params)) {
            executionReq.Auth.Params[k] = interpolateVars(v as string);
          }
        }

        const res = await ExecuteRequest(executionReq);
        results.push({ req, res });
      }
      setMultiRunResults(results);
      setResponse(null); // Clear single response view
      setActiveRequest(null);
    } catch (e) {
      console.error("Multi-run error", e);
      setResponse(new models.APIResponse({
        StatusCode: 0,
        Status: "Error",
        Body: String(e),
        Latency: 0,
        Headers: {},
        ContentType: "raw",
      }));
    } finally {
      setIsMultiRunning(false);
    }
  };


  const handleExportMultiRunResults = async () => {
    if (!multiRunResults) return;
    
    try {
      const filePath = await PromptSaveFilePath("Save Multi-Run Results", "anjal-results.txt");
      if (!filePath) return;
      
      const isTxtOrMd = filePath.endsWith('.txt') || filePath.endsWith('.md');
      
      let content = "";
      
      if (isTxtOrMd) {
        let lines: string[] = [];
        multiRunResults.forEach((r, i) => {
          lines.push(`─────────────────────────────────────────────`);
          lines.push(`Request ${i + 1}: ${r.req.Method} ${r.req.URL}`);
          if (r.req.Auth && r.req.Auth.Type !== "none") {
            lines.push(`   Auth: ${r.req.Auth.Type}`);
          }
          lines.push(`   ✅ Status: ${r.res.Status}`);
          lines.push(`   ⏱  Latency: ${(r.res.Latency / 1000000000).toFixed(2)}s`);
          
          let contentType = r.res.Headers ? (r.res.Headers['Content-Type'] || ['raw'])[0] : "raw";
          lines.push(`   📦 Content-Type: ${contentType} (${r.res.ContentType})`);
          
          if (r.res.Headers) {
            lines.push(`   📋 Headers:`);
            for (const [hk, hv] of Object.entries(r.res.Headers)) {
              lines.push(`      ${hk}: ${(hv as string[]).join(', ')}`);
            }
          }
          
          let preview = r.res.Body || "";
          if (preview.length > 500) {
            preview = preview.substring(0, 500) + "...";
          }
          lines.push(`   📝 Body (${r.res.Body?.length || 0} bytes):`);
          let indentedPreview = preview.split('\n').map(line => `   ${line}`).join('\n');
          lines.push(indentedPreview);
          lines.push("");
        });
        content = lines.join("\n");
      } else {
        const exportData = multiRunResults.map(r => {
          let parsedBody: any = r.res.Body;
          if (r.res.ContentType === 'json' || (r.res.Headers && r.res.Headers['Content-Type']?.[0]?.includes('json'))) {
            try {
              parsedBody = JSON.parse(r.res.Body);
            } catch(e) {}
          }
          return {
            Title: r.req.Title,
            Method: r.req.Method,
            URL: r.req.URL,
            Status: r.res.Status,
            Latency_s: r.res.Latency / 1000000000,
            Response_Headers: r.res.Headers,
            Response_Body: parsedBody,
          };
        });
        content = JSON.stringify(exportData, null, 2);
      }

      await WriteFile(filePath, content);
    } catch (e) {
      console.error("Failed to export results", e);
    }
  };

  if (!isWorkspaceSelected) {
    return (
      <SplashScreen 
        onSelectWorkspace={handleSelectWorkspace}
        onOpenSingleFile={handleOpenSingleFile}
        onUseDefaultWorkspace={() => {
          localStorage.setItem('anjal-workspace', 'default');
          setWorkspacePath('');
          setIsSingleFileMode(false);
          setIsWorkspaceSelected(true);
        }}
      />
    );
  }

  return (
    <div className="flex h-screen w-full bg-[#111111] text-[#F5F4F0] selection:bg-[#E84E38] selection:text-white overflow-hidden" style={{ fontFamily: 'Inter, sans-serif' }}>
      
      <EnvironmentModal 
        isOpen={envModalOpen} 
        onClose={() => setEnvModalOpen(false)} 
        env={envConfig}
        setEnv={handleSaveEnv}
      />

      <Sidebar 
        isOpen={sidebarOpen}
        collections={collections}
        activeRequest={activeRequest}
        isSingleFileMode={isSingleFileMode}
        onOpenCollection={() => {}} // Removed since we use workspace now
        onSelectWorkspace={handleSelectWorkspace}
        onCloseWorkspace={handleCloseWorkspace}
        onSelectRequest={handleSelectRequest}
        onAddRequest={handleAddRequest}
        onAddCollection={handleAddCollection}
        onDeleteRequest={handleDeleteRequest}
        onDeleteCollection={handleDeleteCollection}
        onRunAll={handleExecuteAll}
        isMultiRunning={isMultiRunning}
        onOpenSettings={() => setEnvModalOpen(true)}
      />

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 bg-[#111111]">
        
        <TopUrlBar 
          title={title}
          setTitle={setTitle}
          method={method}
          setMethod={setMethod}
          url={url}
          setUrl={setUrl}
          sidebarOpen={sidebarOpen}
          setSidebarOpen={setSidebarOpen}
          onExecute={handleExecute}
          isExecuting={isExecuting}
          activeRequest={activeRequest}
        />

        {/* Split View Container */}
        <div className="flex-1 flex flex-col lg:flex-row overflow-hidden">
          {multiRunResults ? (
            <div className="flex-1 p-6 overflow-y-auto custom-scrollbar">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-['Playfair_Display'] italic text-[#F5F4F0]">Multi-Run Results</h2>
                <div className="flex gap-4">
                  <button onClick={handleExportMultiRunResults} className="px-4 py-2 border border-[#444] text-sm uppercase tracking-widest hover:text-[#5BA76B] hover:border-[#5BA76B] transition-colors">Export Results</button>
                  <button onClick={() => setMultiRunResults(null)} className="px-4 py-2 border border-[#444] text-sm uppercase tracking-widest hover:bg-[#444] transition-colors">Close</button>
                </div>
              </div>
              <div className="space-y-4">
                {multiRunResults.map((result, idx) => (
                  <MultiRunResultItem key={idx} result={result} />
                ))}
              </div>
            </div>
          ) : (
            <>
              <RequestBox 
                key={activeRequest?.ID || 'none'}
                body={body} setBody={setBody}
                headers={headers} setHeaders={setHeaders}
                params={params} setParams={setParams}
                auth={auth} setAuth={setAuth}
              />
              <ResponseBox response={response} />
            </>
          )}
        </div>
        
        {/* Footer StatusBar */}
        <footer className="border-t border-[#444] bg-[#222] flex items-center px-6 py-2 justify-between text-xs font-['JetBrains_Mono'] text-[#F5F4F0]">
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-[#5BA76B]"></div> Wails Engine Online</span>
            {workspacePath && (
              <span className="text-[#888]">Workspace: {workspacePath}</span>
            )}
          </div>
        </footer>
      </main>

    </div>
  );
}

function MultiRunResultItem({ result }: { result: {req: models.APIRequest, res: models.APIResponse} }) {
  const [expanded, setExpanded] = React.useState(false);
  const success = result.res.StatusCode >= 200 && result.res.StatusCode < 300;

  return (
    <div className="border border-[#444] bg-[#1A1A1A] transition-all">
      <div 
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-[#222]"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-4">
          <span className={`font-['JetBrains_Mono'] font-bold text-sm w-12 ${result.req.Method === 'GET' ? 'text-[#5BA76B]' : result.req.Method === 'POST' ? 'text-[#E09A38]' : result.req.Method === 'PUT' ? 'text-[#3B82F6]' : result.req.Method === 'PATCH' ? 'text-[#A855F7]' : result.req.Method === 'DELETE' ? 'text-[#E84E38]' : 'text-[#888]'}`}>{result.req.Method}</span>
          <span className="text-[#F5F4F0] font-medium">{result.req.URL}</span>
        </div>
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-1 text-xs text-[#888] font-['JetBrains_Mono']">
            <Clock className="w-4 h-4" />
            <span>{(result.res.Latency / 1000000000).toFixed(2)}s</span>
          </div>
          <span className={`px-3 py-1 text-xs font-['JetBrains_Mono'] font-bold ${success ? 'text-[#5BA76B] bg-[#5BA76B]/10' : 'text-[#E84E38] bg-[#E84E38]/10'}`}>
            {result.res.Status}
          </span>
        </div>
      </div>
      
      {expanded && (
        <div className="border-t border-[#444] p-4 bg-[#111111] max-h-96 overflow-auto text-xs font-['JetBrains_Mono'] text-[#F5F4F0]">
          
          {result.req.Headers && Object.keys(result.req.Headers).length > 0 && (
            <div className="mb-4">
              <div className="text-[#888] mb-1 font-bold uppercase tracking-widest text-[10px]">Request Headers</div>
              <div className="bg-[#1A1A1A] p-2 border border-[#444]">
                {Object.entries(result.req.Headers).map(([k, v]) => (
                  <div key={k}><span className="text-[#5BA76B]">{k}:</span> {v}</div>
                ))}
              </div>
            </div>
          )}

          {result.res.Headers && Object.keys(result.res.Headers).length > 0 && (
            <div className="mb-4">
              <div className="text-[#888] mb-1 font-bold uppercase tracking-widest text-[10px]">Response Headers</div>
              <div className="bg-[#1A1A1A] p-2 border border-[#444]">
                {Object.entries(result.res.Headers).map(([k, v]) => (
                  <div key={k}><span className="text-[#5BA76B]">{k}:</span> {(v as string[]).join(', ')}</div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="text-[#888] mb-1 font-bold uppercase tracking-widest text-[10px]">Response Body</div>
            <pre className="bg-[#1A1A1A] p-4 border border-[#444] whitespace-pre-wrap">
              {(() => {
                if (!result.res.Body) return '<No Content>';
                if (result.res.ContentType === 'json' || (result.res.Headers && result.res.Headers['Content-Type']?.[0]?.includes('json'))) {
                  try {
                    return JSON.stringify(JSON.parse(result.res.Body), null, 2);
                  } catch (e) {
                    return result.res.Body;
                  }
                }
                return result.res.Body;
              })()}
            </pre>
          </div>
          
        </div>
      )}
    </div>
  );
}
