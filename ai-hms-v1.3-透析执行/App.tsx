
import React, { useState } from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import DialysisExecution from './pages/DialysisExecution';

const App: React.FC = () => {
  const [isSidebarVisible, setIsSidebarVisible] = useState(true);

  return (
    <HashRouter>
      <div className="flex h-screen w-full bg-slate-50 overflow-hidden">
        <div className={`transition-all duration-300 ease-in-out ${isSidebarVisible ? 'w-52' : 'w-0'}`}>
           <Sidebar isVisible={isSidebarVisible} />
        </div>
        <div className="flex flex-col flex-1 overflow-hidden min-w-0">
          <Header onToggleSidebar={() => setIsSidebarVisible(!isSidebarVisible)} isSidebarVisible={isSidebarVisible} />
          <main className="flex-1 overflow-hidden">
            <Routes>
              <Route path="/execution" element={<DialysisExecution />} />
              <Route path="/" element={<Navigate to="/execution" replace />} />
            </Routes>
          </main>
        </div>
      </div>
    </HashRouter>
  );
};

export default App;
