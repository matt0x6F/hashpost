'use client';

import { useState, useEffect } from 'react';
import { ScrollArea } from '@/components/shadcn/scroll-area';
import { ChevronDown, ChevronRight, Shield, AlertTriangle, Info, Ban } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { type Rule } from './RulesEditor';

interface PlatformRulesDisplayProps {
  showTitle?: boolean;
  maxHeight?: string;
  compact?: boolean;
}

export function PlatformRulesDisplay({ 
  showTitle = true, 
  maxHeight = "h-64",
  compact = false 
}: PlatformRulesDisplayProps) {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedRules, setExpandedRules] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadPlatformRules();
  }, []);

  const loadPlatformRules = async () => {
    try {
      setLoading(true);
      // Platform rules not available in atproto system
      setRules([]);
    } catch (error) {
      console.error('Error loading platform rules:', error);
      // Don't show error toast for public display
    } finally {
      setLoading(false);
    }
  };

  const toggleRule = (ruleCode: string) => {
    const newExpanded = new Set(expandedRules);
    if (newExpanded.has(ruleCode)) {
      newExpanded.delete(ruleCode);
    } else {
      newExpanded.add(ruleCode);
    }
    setExpandedRules(newExpanded);
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <Ban className="w-4 h-4 text-destructive" />;
      case 'high':
        return <AlertTriangle className="w-4 h-4 text-orange-500" />;
      case 'medium':
        return <Shield className="w-4 h-4 text-yellow-500" />;
      case 'low':
        return <Info className="w-4 h-4 text-blue-500" />;
      default:
        return <Info className="w-4 h-4 text-muted-foreground" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'border-destructive/20 bg-destructive/5';
      case 'high':
        return 'border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950/20';
      case 'medium':
        return 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950/20';
      case 'low':
        return 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950/20';
      default:
        return 'border-border bg-card';
    }
  };

  if (loading) {
    return (
      <div className="space-y-3">
        {showTitle && (
          <h3 className="text-sm font-semibold text-foreground">Platform Rules</h3>
        )}
        <div className="text-xs text-muted-foreground">Loading platform rules...</div>
      </div>
    );
  }

  if (rules.length === 0) {
    return (
      <div className="space-y-3">
        {showTitle && (
          <h3 className="text-sm font-semibold text-foreground">Platform Rules</h3>
        )}
        <div className="text-xs text-muted-foreground">No platform rules configured</div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {showTitle && (
        <h3 className="text-sm font-semibold text-foreground">Platform Rules</h3>
      )}
      <ScrollArea className={maxHeight}>
        <div className="space-y-3 pr-4">
          {rules.map((rule) => {
            const isExpanded = expandedRules.has(rule.code);
            return (
              <div 
                key={rule.code} 
                className={`border rounded-lg transition-colors group ${getSeverityColor(rule.severity)} hover:bg-muted/50`}
              >
                <button
                  onClick={() => toggleRule(rule.code)}
                  className="w-full flex items-center justify-between p-3 text-left"
                >
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    {getSeverityIcon(rule.severity)}
                    <h4 className={`font-medium text-foreground truncate ${compact ? 'text-xs' : 'text-sm'}`}>
                      {rule.name}
                    </h4>
                    {!compact && (
                      <span className="text-xs text-muted-foreground capitalize">
                        {rule.category}
                      </span>
                    )}
                  </div>
                  {isExpanded ? (
                    <ChevronDown className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                  ) : (
                    <ChevronRight className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                  )}
                </button>
                {isExpanded && (
                  <div className="px-3 pb-3">
                    <p className={`text-muted-foreground group-hover:text-foreground/80 ${compact ? 'text-xs' : 'text-sm'}`}>
                      {rule.description}
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
