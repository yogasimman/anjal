import React, { useState, useEffect } from 'react';
import { 
  GetCollections, 
  GetCollectionsFrom, 
  PromptOpenWorkspace, 
  ExecuteRequest,
  CreateRequest,
  UpdateRequest,
  DeleteRequest,
  DeleteCollection
} from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';

import Sidebar from './components/layout/Sidebar';
import TopUrlBar from './components/layout/TopUrlBar';
import RequestBox from './components/layout/RequestBox';
import ResponseBox from './components/layout/ResponseBox';
import EnvironmentModal from './components/layout/EnvironmentModal';

export default function App() {
  const [method, setMethod] = useState('GET');
  const [url, setUrl] = useState('https://api.anjal.dev/v1/users');
  const [sidebarOpen, setSidebarOpen] = useState(true);

  // App State
  const [workspacePath, setWorkspacePath] = useState<string>('');
  const [collections, setCollections] = useState<models.Collection[]>([]);
  const [activeRequest, setActiveRequest] = useState<models.APIRequest | null>(null);
  const [activeCollection, setActiveCollection] = useState<models.Collection | null>(null);
  const [response, setResponse] = useState<models.APIResponse | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);

  // Environment & Local Config
  const [envModalOpen, setEnvModalOpen] = useState(false);
  const [envConfig, setEnvConfig] = useState<Record<string, string>>(() => {
    const saved = localStorage.getItem('anjal-env');
    return saved ? JSON.parse(saved) : { "BASE_URL": "https://api.example.com" };
  });

  useEffect(() => {
    localStorage.setItem('anjal-env', JSON.stringify(envConfig));
  }, [envConfig]);

  useEffect(() => {
    if (workspacePath) {
      loadCollectionsFrom(workspacePath);
    } else {
      loadCollections();
    }
  }, [workspacePath]);

  const loadCollections = async () => {
    try {
      const cols = await GetCollections();
      if (cols) setCollections(cols);
    } catch (e) {
      console.error("Failed to load collections", e);
    }
  };

  const loadCollectionsFrom = async (dir: string) => {
    try {
      const cols = await GetCollectionsFrom(dir);
      if (cols) setCollections(cols);
    } catch (e) {
      console.error("Failed to load collections from", dir, e);
    }
  };

  const handleSelectWorkspace = async () => {
    try {
      const dir = await PromptOpenWorkspace();
      if (dir) {
        setWorkspacePath(dir);
      }
    } catch (e) {
      console.error("Failed to open workspace", e);
    }
  };

  const handleSelectRequest = (req: models.APIRequest) => {
    setActiveRequest(req);
    // Find parent collection for saving updates later
    const col = collections.find(c => c.Requests?.some(r => r.ID === req.ID));
    if (col) setActiveCollection(col);

    setMethod(req.Method || 'GET');
    setUrl(req.URL || '');
    setResponse(null);
  };

  // CRUD Operations
  const handleAddRequest = async (col: models.Collection) => {
    const newReq = new models.APIRequest({
      Title: "New Request",
      Method: "GET",
      URL: "https://",
    });
    try {
      await CreateRequest(col.FilePath, newReq);
      // Reload collections
      if (workspacePath) loadCollectionsFrom(workspacePath);
      else loadCollections();
    } catch (e) {
      console.error("Failed to create request", e);
    }
  };

  const handleDeleteRequest = async (col: models.Collection, req: models.APIRequest) => {
    try {
      await DeleteRequest(col.FilePath, req.ID);
      if (activeRequest?.ID === req.ID) {
        setActiveRequest(null);
        setResponse(null);
      }
      if (workspacePath) loadCollectionsFrom(workspacePath);
      else loadCollections();
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
      if (workspacePath) loadCollectionsFrom(workspacePath);
      else loadCollections();
    } catch (e) {
      console.error("Failed to delete collection", e);
    }
  };

  // Execution & Variable Interpolation
  const interpolateVars = (str: string) => {
    let result = str;
    for (const [key, value] of Object.entries(envConfig)) {
      result = result.replace(new RegExp(`{{${key}}}`, 'g'), value);
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
        Method: method,
        URL: url,
      });

      if (activeCollection) {
        await UpdateRequest(activeCollection.FilePath, activeRequest.ID, reqToExecute);
      }

      // 2. Interpolate variables for execution (don't save interpolated back to MD)
      const executionReq = new models.APIRequest({
        ...reqToExecute,
        URL: interpolateVars(reqToExecute.URL),
        Body: interpolateVars(reqToExecute.Body || ''),
      });

      const res = await ExecuteRequest(executionReq);
      setResponse(res);
      
      // Reload to reflect saved changes in sidebar
      if (workspacePath) loadCollectionsFrom(workspacePath);
      else loadCollections();

    } catch (e) {
      console.error(e);
    } finally {
      setIsExecuting(false);
    }
  };

  return (
    <div className="flex h-screen w-full bg-[#111111] text-[#F5F4F0] selection:bg-[#E84E38] selection:text-white overflow-hidden" style={{ fontFamily: 'Inter, sans-serif' }}>
      
      <EnvironmentModal 
        isOpen={envModalOpen} 
        onClose={() => setEnvModalOpen(false)} 
        env={envConfig}
        setEnv={setEnvConfig}
      />

      <Sidebar 
        isOpen={sidebarOpen}
        collections={collections}
        activeRequest={activeRequest}
        onOpenCollection={() => {}} // Removed since we use workspace now
        onSelectWorkspace={handleSelectWorkspace}
        onSelectRequest={handleSelectRequest}
        onAddRequest={handleAddRequest}
        onDeleteRequest={handleDeleteRequest}
        onDeleteCollection={handleDeleteCollection}
        onOpenSettings={() => setEnvModalOpen(true)}
      />

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 bg-[#111111]">
        
        <TopUrlBar 
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
          <RequestBox activeRequest={activeRequest} />
          <ResponseBox response={response} />
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
