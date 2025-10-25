"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Switch } from "@/components/shadcn/switch";
import { Badge } from "@/components/shadcn/badge";
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import { 
  Shield, 
  Database,
  Key,
  Bell,
  Save,
  RefreshCw
} from "lucide-react";
import { toast } from "sonner";

interface SystemSetting {
  key: string;
  value: string | boolean | number;
  type: "string" | "boolean" | "number";
  category: string;
  description: string;
  requiresRestart?: boolean;
}

const mockSettings: SystemSetting[] = [
  {
    key: "platform_name",
    value: "HashPost",
    type: "string",
    category: "General",
    description: "Display name for the platform"
  },
  {
    key: "maintenance_mode",
    value: false,
    type: "boolean",
    category: "System",
    description: "Enable maintenance mode for all users",
    requiresRestart: true
  },
  {
    key: "content_moderation_enabled",
    value: true,
    type: "boolean",
    category: "Moderation",
    description: "Enable automatic content moderation"
  },
  {
    key: "registration_enabled",
    value: true,
    type: "boolean",
    category: "Users",
    description: "Allow new user registrations"
  },
  {
    key: "max_post_length",
    value: 10000,
    type: "number",
    category: "Content",
    description: "Maximum length for posts in characters"
  },
  {
    key: "session_timeout_hours",
    value: 24,
    type: "number",
    category: "Security",
    description: "User session timeout in hours"
  }
];

export function SystemSettingsTab() {
  const [settings, setSettings] = useState<SystemSetting[]>(mockSettings);
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const categories = Array.from(new Set(settings.map(s => s.category)));

  const updateSetting = (key: string, value: string | boolean | number) => {
    setSettings(settings.map(s => 
      s.key === key ? { ...s, value } : s
    ));
    setHasUnsavedChanges(true);
  };

  const saveSettings = async () => {
    try {
      setIsLoading(true);
      // In atproto system, system settings are managed differently
      toast.error("System settings are not available in the atproto system");
    } catch (error) {
      console.error("Failed to save settings:", error);
      toast.error("Failed to save settings");
    } finally {
      setIsLoading(false);
    }
  };

  const filteredSettings = settings.filter(setting => {
    const matchesCategory = categoryFilter === "all" || setting.category === categoryFilter;
    const matchesSearch = setting.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         setting.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  return (
    <div className="space-y-6">
      {/* System Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            System Settings
          </CardTitle>
          <CardDescription>
            Configure platform-wide settings and preferences
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex gap-4">
              <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                <SelectTrigger className="w-48">
                  <SelectValue placeholder="Filter by category" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Categories</SelectItem>
                  {categories.map(category => (
                    <SelectItem key={category} value={category}>{category}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              
              <Input
                placeholder="Search settings..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="flex-1"
              />
            </div>

            <div className="space-y-4">
              {filteredSettings.map((setting) => (
                <div key={setting.key} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium">{setting.key}</span>
                      <Badge variant="outline">{setting.category}</Badge>
                      {setting.requiresRestart && (
                        <Badge variant="destructive">Requires Restart</Badge>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">{setting.description}</p>
                  </div>
                  
                  <div className="ml-4">
                    {setting.type === "boolean" ? (
                      <Switch
                        checked={setting.value as boolean}
                        onCheckedChange={(checked) => updateSetting(setting.key, checked)}
                      />
                    ) : setting.type === "number" ? (
                      <Input
                        type="number"
                        value={setting.value as number}
                        onChange={(e) => updateSetting(setting.key, parseInt(e.target.value) || 0)}
                        className="w-24"
                      />
                    ) : (
                      <Input
                        value={setting.value as string}
                        onChange={(e) => updateSetting(setting.key, e.target.value)}
                        className="w-48"
                      />
                    )}
                  </div>
                </div>
              ))}
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => window.location.reload()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Reset
              </Button>
              <Button 
                onClick={saveSettings} 
                disabled={!hasUnsavedChanges || isLoading}
                className="min-w-[120px]"
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin h-4 w-4 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Saving...
                  </>
                ) : (
                  <>
                    <Save className="h-4 w-4 mr-2" />
                    {hasUnsavedChanges ? "Save Changes" : "Saved"}
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}