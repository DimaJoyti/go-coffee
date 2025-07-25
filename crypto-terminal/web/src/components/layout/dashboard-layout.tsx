import * as React from "react"
import { cn } from "../../lib/utils"
import { Sidebar } from "./sidebar"
import { Header } from "./header"
import { Button } from "../ui/button"
import { ScrollArea } from "../ui/scroll-area"
import { Toaster } from "../ui/toaster"
import { Menu, X } from "lucide-react"

interface DashboardLayoutProps {
  children: React.ReactNode
  className?: string
}

const DashboardLayout: React.FC<DashboardLayoutProps> = ({
  children,
  className,
}) => {
  const [sidebarCollapsed, setSidebarCollapsed] = React.useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false)
  const [isDark, setIsDark] = React.useState(false)
  const [connectionStatus, setConnectionStatus] = React.useState<"connected" | "disconnected" | "connecting">("connected")

  // Handle theme toggle
  const handleThemeToggle = () => {
    setIsDark(!isDark)
    // In a real app, you'd update the theme context or localStorage
    document.documentElement.classList.toggle("dark")
  }

  // Handle sidebar toggle
  const handleSidebarToggle = () => {
    setSidebarCollapsed(!sidebarCollapsed)
  }

  // Handle mobile menu toggle
  const handleMobileMenuToggle = () => {
    setMobileMenuOpen(!mobileMenuOpen)
  }

  // Close mobile menu when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element
      if (mobileMenuOpen && !target.closest(".mobile-sidebar") && !target.closest(".mobile-menu-button")) {
        setMobileMenuOpen(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [mobileMenuOpen])

  // Simulate connection status changes
  React.useEffect(() => {
    const interval = setInterval(() => {
      // Randomly change connection status for demo
      const statuses: Array<"connected" | "disconnected" | "connecting"> = ["connected", "connected", "connected", "connecting"]
      const randomStatus = statuses[Math.floor(Math.random() * statuses.length)]
      setConnectionStatus(randomStatus)
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  return (
    <div className={cn("flex h-screen bg-background", className)}>
      {/* Mobile Menu Button */}
      <Button
        variant="ghost"
        size="icon"
        className="mobile-menu-button fixed top-4 left-4 z-50 md:hidden"
        onClick={handleMobileMenuToggle}
      >
        {mobileMenuOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
      </Button>

      {/* Mobile Sidebar Overlay */}
      {mobileMenuOpen && (
        <div className="fixed inset-0 z-40 bg-background/80 backdrop-blur-sm md:hidden" />
      )}

      {/* Sidebar */}
      <aside className={cn(
        "mobile-sidebar fixed inset-y-0 left-0 z-50 transition-transform duration-300 md:relative md:translate-x-0",
        mobileMenuOpen ? "translate-x-0" : "-translate-x-full",
        "md:block"
      )}>
        <Sidebar
          collapsed={sidebarCollapsed}
          onToggle={handleSidebarToggle}
          className="h-full"
        />
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Header */}
        <Header
          onThemeToggle={handleThemeToggle}
          isDark={isDark}
          connectionStatus={connectionStatus}
          notifications={3}
        />

        {/* Main Content Area */}
        <main className="flex-1 overflow-hidden">
          <ScrollArea className="h-full">
            <div className="p-6 space-y-6">
              {children}
            </div>
          </ScrollArea>
        </main>
      </div>

      {/* Toast Notifications */}
      <Toaster />
    </div>
  )
}

export { DashboardLayout }
