'use client';

import { useState, useEffect } from 'react';
import { ScrollArea } from '@/components/shadcn/scroll-area';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { type Rule } from './RulesEditor';

interface SubforumRulesDisplayProps {
  communityType: string;
  subforumName: string;
}

export function SubforumRulesDisplay({ communityType, subforumName }: SubforumRulesDisplayProps) {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedRules, setExpandedRules] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadRules();
  }, [communityType, subforumName]);

  const loadRules = async () => {
    try {
      setLoading(true);
      // Subforum rules not available in atproto system
      setRules([]);
    } catch (error) {
      console.error('Error loading rules:', error);
      // Don't show error toast for sidebar display
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

  if (loading) {
    return (
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-foreground">Community Rules</h3>
        <div className="text-xs text-muted-foreground">Loading rules...</div>
      </div>
    );
  }

  if (rules.length === 0) {
    return (
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-foreground">Community Rules</h3>
        <div className="text-xs text-muted-foreground">No rules configured</div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-foreground">Community Rules</h3>
      <ScrollArea className="h-64">
        <div className="space-y-3 pr-4">
          {rules.map((rule) => {
            const isExpanded = expandedRules.has(rule.code);
            return (
              <div key={rule.code} className="border rounded-lg bg-card hover:bg-muted/50 transition-colors group">
                <button
                  onClick={() => toggleRule(rule.code)}
                  className="w-full flex items-center justify-between p-3 text-left"
                >
                  <h4 className="text-xs font-medium text-foreground">{rule.name}</h4>
                  {isExpanded ? (
                    <ChevronDown className="w-4 h-4 text-muted-foreground" />
                  ) : (
                    <ChevronRight className="w-4 h-4 text-muted-foreground" />
                  )}
                </button>
                {isExpanded && (
                  <div className="px-3 pb-3">
                    <p className="text-xs text-muted-foreground group-hover:text-foreground/80">{rule.description}</p>
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