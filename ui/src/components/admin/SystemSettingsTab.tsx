"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Textarea } from "@/components/shadcn/textarea";
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
  Settings, 
  Shield, 
  Users, 
  Globe,
  Database,
  Key,
  Bell,
  Save,
  RefreshCw
} from "lucide-react";
import { toast } from "sonner";
import { getApi } from "@/lib/api-client";
import { RulesApi } from "@/generated/api/src/apis/RulesApi";
import { Rule } from "@/generated/api/src/models/Rule";

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
    key: "max_pseudonyms_per_user",
    value: 5,
    type: "number",
    category: "Users",
    description: "Maximum number of pseudonyms a user can create"
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
  },
  {
    key: "correlation_audit_enabled",
    value: true,
    type: "boolean",
    category: "Privacy",
    description: "Log all identity correlation requests for audit"
  }
];

export function SystemSettingsTab() {
  const [settings, setSettings] = useState<SystemSetting[]>(mockSettings);
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  
  // Platform rules state
  const [platformRules, setPlatformRules] = useState<Rule[]>([]);
  const [contentGuidelines, setContentGuidelines] = useState("");
  const [hasUnsavedRules, setHasUnsavedRules] = useState(false);

  // Load platform rules on component mount
  useEffect(() => {
    loadPlatformRules();
  }, []);

  const loadPlatformRules = async () => {
    try {
      const rulesApi = getApi(RulesApi);
      const response = await rulesApi.getPlatformRules(false);
      setPlatformRules(response.rules || []);
    } catch (error) {
      console.error('Error loading platform rules:', error);
      // Don't show error toast for initial load
    }
  };

  const savePlatformRules = async () => {
    try {
      setIsLoading(true);
      const rulesApi = getApi(RulesApi);
      
      // Update platform rules
      await rulesApi.updatePlatformRules({
        rules: platformRules
      });
      
      // TODO: Add content guidelines endpoint when available
      // For now, just save the rules
      
      toast.success("Platform rules saved successfully");
      setHasUnsavedRules(false);
    } catch (error: any) {
      console.error("Failed to save platform rules:", error);
      toast.error("Failed to save platform rules. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const categories = Array.from(new Set(settings.map(s => s.category)));

  const updateSetting = (key: string, value: any) => {
    setSettings(settings.map(s => 
      s.key === key ? { ...s, value } : s
    ));
    setHasUnsavedChanges(true);
  };

  const saveSettings = async () => {
    try {
      setIsLoading(true);
      
      // TODO: Implement actual API call when system settings update endpoint is available
      // For now, simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      toast.success("Settings saved successfully");
      setHasUnsavedChanges(false);
    } catch (error: any) {
      console.error("Failed to save settings:", error);
      toast.error("Failed to save settings. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const resetSettings = () => {
    setSettings(mockSettings);
    setHasUnsavedChanges(false);
    toast.info("Settings reset to defaults");
  };

  const filteredSettings = settings.filter(setting => {
    if (categoryFilter !== "all" && setting.category !== categoryFilter) return false;
    if (searchQuery && !setting.key.toLowerCase().includes(searchQuery.toLowerCase()) && 
        !setting.description.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    return true;
  });

  const renderSettingValue = (setting: SystemSetting) => {
    switch (setting.type) {
      case "boolean":
        return (
          <Switch
            checked={setting.value as boolean}
            onCheckedChange={(checked) => updateSetting(setting.key, checked)}
          />
        );
      case "number":
        return (
          <Input
            type="number"
            value={setting.value as number}
            onChange={(e) => updateSetting(setting.key, parseInt(e.target.value) || 0)}
            className="w-32"
          />
        );
      case "string":
        return (
          <Input
            value={setting.value as string}
            onChange={(e) => updateSetting(setting.key, e.target.value)}
            className="w-64"
          />
        );
      default:
        return <span>{String(setting.value)}</span>;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Actions */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">System Settings</h2>
          <p className="text-muted-foreground">
            Configure platform-wide settings and behavior
          </p>
        </div>
        
        <div className="flex gap-2">
          <Button variant="outline" onClick={resetSettings}>
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

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Category</label>
              <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="All Categories" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Categories</SelectItem>
                  {categories.map(category => (
                    <SelectItem key={category} value={category}>{category}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Search</label>
              <Input
                placeholder="Search settings..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Settings Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        {filteredSettings.map((setting) => (
          <Card key={setting.key}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span className="text-sm font-medium">{setting.key}</span>
                {setting.requiresRestart && (
                  <Badge variant="destructive" className="text-xs">Restart Required</Badge>
                )}
              </CardTitle>
              <CardDescription>{setting.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {renderSettingValue(setting)}
                </div>
                <Badge variant="outline" className="text-xs">
                  {setting.category}
                </Badge>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>
            Common system administration tasks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Database className="h-6 w-6" />
              <span>Database Backup</span>
            </Button>
            
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Key className="h-6 w-6" />
              <span>Rotate Keys</span>
            </Button>
            
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Bell className="h-6 w-6" />
              <span>Send Announcement</span>
            </Button>
            
            <Button variant="outline" className="h-20 flex-col gap-2">
              <Shield className="h-6 w-6" />
              <span>Security Audit</span>
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Platform Rules */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            Platform Rules
          </CardTitle>
          <CardDescription>
            Global rules that apply to all subforums
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Platform Rules</label>
              <div className="space-y-2">
                {platformRules.map((rule, index) => (
                  <div key={index} className="flex items-center gap-2 p-2 border rounded">
                    <Input
                      placeholder="Rule code (e.g., no_harassment)"
                      value={rule.code}
                      onChange={(e) => {
                        const newRules = [...platformRules];
                        newRules[index] = { ...rule, code: e.target.value };
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                      className="w-32"
                    />
                    <Input
                      placeholder="Rule name"
                      value={rule.name}
                      onChange={(e) => {
                        const newRules = [...platformRules];
                        newRules[index] = { ...rule, name: e.target.value };
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                      className="flex-1"
                    />
                    <Select
                      value={rule.category}
                      onValueChange={(value) => {
                        const newRules = [...platformRules];
                        newRules[index] = { ...rule, category: value };
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                    >
                      <SelectTrigger className="w-32">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="content">Content</SelectItem>
                        <SelectItem value="behavior">Behavior</SelectItem>
                        <SelectItem value="safety">Safety</SelectItem>
                        <SelectItem value="spam">Spam</SelectItem>
                        <SelectItem value="legal">Legal</SelectItem>
                      </SelectContent>
                    </Select>
                    <Select
                      value={rule.severity}
                      onValueChange={(value) => {
                        const newRules = [...platformRules];
                        newRules[index] = { ...rule, severity: value };
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                    >
                      <SelectTrigger className="w-24">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="low">Low</SelectItem>
                        <SelectItem value="medium">Medium</SelectItem>
                        <SelectItem value="high">High</SelectItem>
                        <SelectItem value="critical">Critical</SelectItem>
                      </SelectContent>
                    </Select>
                    <Switch
                      checked={rule.active}
                      onCheckedChange={(checked) => {
                        const newRules = [...platformRules];
                        newRules[index] = { ...rule, active: checked };
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        const newRules = platformRules.filter((_, i) => i !== index);
                        setPlatformRules(newRules);
                        setHasUnsavedRules(true);
                      }}
                    >
                      Remove
                    </Button>
                  </div>
                ))}
                <Button
                  variant="outline"
                  onClick={() => {
                    setPlatformRules([...platformRules, {
                      code: '',
                      name: '',
                      description: '',
                      category: 'content',
                      severity: 'medium',
                      active: true,
                    }]);
                    setHasUnsavedRules(true);
                  }}
                >
                  Add Rule
                </Button>
              </div>
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Content Guidelines</label>
              <Textarea
                placeholder="Enter content guidelines..."
                rows={4}
                className="w-full"
                value={contentGuidelines}
                onChange={(e) => {
                  setContentGuidelines(e.target.value);
                  setHasUnsavedRules(true);
                }}
              />
            </div>
            
            <Button 
              onClick={savePlatformRules} 
              disabled={!hasUnsavedRules || isLoading}
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
                  {hasUnsavedRules ? "Save Rules" : "Saved"}
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
