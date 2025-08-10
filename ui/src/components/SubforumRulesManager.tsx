'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { RulesApi } from '@/generated/api/src/apis';
import { RulesEditor, type Rule } from './RulesEditor';



interface SubforumRulesManagerProps {
  communityType: string;
  subforumName: string;
  subforumId: number;
}

export function SubforumRulesManager({ communityType, subforumName, subforumId }: SubforumRulesManagerProps) {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadRules();
  }, [subforumId]);

  const loadRules = async () => {
    try {
      setLoading(true);
      const rulesApi = getApi(RulesApi);
      const response = await rulesApi.getSubforumRulesRaw({
        communityType,
        subforumName,
        activeOnly: false
      });
      
      if (response.raw.status === 200) {
        const data = await response.value();
        setRules(data.rules || []);
      } else {
        toast.error('Failed to load rules');
      }
    } catch (error) {
      console.error('Error loading rules:', error);
      toast.error('Failed to load rules');
    } finally {
      setLoading(false);
    }
  };

  const handleRulesChange = async (newRules: Rule[]) => {
    setRules(newRules);
    await saveRules(newRules);
  };

  const saveRules = async (rulesToSave: Rule[]) => {
    try {
      setSaving(true);
      
      // Get current rules from backend
      const rulesApi = getApi(RulesApi);
      const currentResponse = await rulesApi.getSubforumRulesRaw({
        communityType,
        subforumName,
        activeOnly: false
      });
      
      const currentRules = currentResponse.raw.status === 200 ? (await currentResponse.value()).rules || [] : [];
      
      // Find rules to create, update, and delete
      const currentRuleCodes = new Set(currentRules.map(r => r.code));
      const newRuleCodes = new Set(rulesToSave.map(r => r.code));
      
      // Delete rules that are no longer present
      for (const currentRule of currentRules) {
        if (!newRuleCodes.has(currentRule.code)) {
          await rulesApi.deleteSubforumRule(communityType, subforumName, currentRule.code);
        }
      }
      
      // Create or update rules
      for (const rule of rulesToSave) {
        if (currentRuleCodes.has(rule.code)) {
          // Update existing rule
          await rulesApi.updateSubforumRule(
            communityType,
            subforumName,
            rule.code,
            {
              name: rule.name,
              description: rule.description,
              category: rule.category,
              severity: rule.severity,
              active: rule.active,
            }
          );
        } else {
          // Create new rule
          await rulesApi.createSubforumRule(
            communityType,
            subforumName,
            {
              code: rule.code,
              name: rule.name,
              description: rule.description,
              category: rule.category,
              severity: rule.severity,
              active: rule.active,
            }
          );
        }
      }
      
      toast.success('Rules updated successfully');
    } catch (error) {
      console.error('Error saving rules:', error);
      toast.error('Failed to save rules');
      // Reload rules to revert UI state
      loadRules();
    } finally {
      setSaving(false);
    }
  };



  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Rules Management</CardTitle>
          <CardDescription>Configure rules for this subforum</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading rules...</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <RulesEditor
      rules={rules}
      onChange={handleRulesChange}
      disabled={saving}
      showTitle={true}
      maxRules={20}
    />
  );
} 