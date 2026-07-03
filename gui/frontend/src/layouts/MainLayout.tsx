import React, { ReactNode } from 'react';
import Sidebar from '../components/Sidebar';
import Header from '../components/Header';

interface MainLayoutProps {
  children: ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[radial-gradient(circle_at_top_right,_var(--tw-gradient-stops))] from-slate-800 to-slate-900">
      <Sidebar />
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        <Header />
        <main className="flex-1 p-6 overflow-y-auto relative">
          {children}
        </main>
      </div>
    </div>
  );
};

export default MainLayout;
