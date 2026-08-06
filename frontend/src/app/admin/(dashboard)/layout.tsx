"use client";

import { useAuth } from '@/context/AuthContext';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LogOut, LayoutDashboard, Users, Map, Shield, Cpu } from 'lucide-react';

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  if (!user) return null;

  return (
    <div className="flex flex-col min-h-screen">
      <header className="nav-header">
        <div className="nav-container">
          <div className="flex items-center gap-2">
            <Shield className="text-accent" size={24} />
            <span className="font-bold text-lg">TraceLock SOC</span>
          </div>

          <nav className="nav-links flex items-center gap-6">
            <Link 
              href="/admin" 
              className={`nav-link flex items-center gap-2 ${pathname === '/admin' ? 'active' : ''}`}
            >
              <LayoutDashboard size={18} />
              <span className="text-sm">Dashboard</span>
            </Link>
            
            <Link 
              href="/admin/zones" 
              className={`nav-link flex items-center gap-2 ${pathname.startsWith('/admin/zones') ? 'active' : ''}`}
            >
              <Map size={18} />
              <span className="text-sm">Zones</span>
            </Link>
            
            <Link 
              href="/admin/users" 
              className={`nav-link flex items-center gap-2 ${pathname.startsWith('/admin/users') ? 'active' : ''}`}
            >
              <Users size={18} />
              <span className="text-sm">Users</span>
            </Link>

            <Link 
              href="/admin/simulator" 
              className={`nav-link flex items-center gap-2 ${pathname.startsWith('/admin/simulator') ? 'active' : ''}`}
            >
              <Cpu size={18} />
              <span className="text-sm">Simulator</span>
            </Link>
          </nav>

          <div className="flex items-center gap-4">
            <div className="flex flex-col items-end">
              <span className="text-sm font-medium">{user.name}</span>
              <span className="text-xs text-accent mono uppercase">{user.role}</span>
            </div>
            
            <button 
              onClick={logout}
              className="btn btn-danger"
              style={{ padding: '0.5rem' }}
              title="Logout"
            >
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </header>

      <main className="container flex-1 w-full mt-4">
        {children}
      </main>
    </div>
  );
}
