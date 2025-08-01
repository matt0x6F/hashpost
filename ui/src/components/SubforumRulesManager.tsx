'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Textarea } from '@/components/shadcn/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/shadcn/select';
import { Switch } from '@/components/shadcn/switch';
import { Badge } from '@/components/shadcn/badge';
import { Separator } from '@/components/shadcn/separator';
import { Plus, Edit, Trash2, Save, X } from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { RulesApi } from '@/generated/api/src/apis';

interface Rule {
  code: string;
  name: string;
  description: string;
  category: string;
  severity: string;
  active: boolean;
}

interface SubforumRulesManagerProps {
  communityType: string;
  subforumName: string;
  subforumId: number;
}

export function SubforumRulesManager({ communityType, subforumName, subforumId }: SubforumRulesManagerProps) {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const categories = [
    { value: 'content', label: 'Content' },
    { value: 'behavior', label: 'Behavior' },
    { value: 'safety', label: 'Safety' },
    { value: 'spam', label: 'Spam' },
    { value: 'legal', label: 'Legal' },
  ];

  const severities = [
    { value: 'low', label: 'Low' },
    { value: 'medium', label: 'Medium' },
    { value: 'high', label: 'High' },
    { value: 'critical', label: 'Critical' },
  ];

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

  const handleCreateRule = () => {
    setEditingRule({
      code: '',
      name: '',
      description: '',
      category: 'content',
      severity: 'medium',
      active: true,
    });
    setIsCreating(true);
  };

  const handleEditRule = (rule: Rule) => {
    setEditingRule({ ...rule });
    setIsCreating(false);
  };

  const handleSaveRule = async () => {
    if (!editingRule) return;

    try {
      setSaving(true);
      const rulesApi = getApi(RulesApi);

      if (isCreating) {
        await rulesApi.createSubforumRule(
          communityType,
          subforumName,
          {
            code: editingRule.code,
            name: editingRule.name,
            description: editingRule.description,
            category: editingRule.category,
            severity: editingRule.severity,
            active: editingRule.active,
          }
        );
        toast.success('Rule created successfully');
      } else {
        await rulesApi.updateSubforumRule(
          communityType,
          subforumName,
          editingRule.code,
          {
            name: editingRule.name,
            description: editingRule.description,
            category: editingRule.category,
            severity: editingRule.severity,
            active: editingRule.active,
          }
        );
        toast.success('Rule updated successfully');
      }

      setEditingRule(null);
      setIsCreating(false);
      loadRules();
    } catch (error) {
      console.error('Error saving rule:', error);
      toast.error('Failed to save rule');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteRule = async (ruleCode: string) => {
    if (!confirm('Are you sure you want to delete this rule?')) return;

    try {
      setSaving(true);
      const rulesApi = getApi(RulesApi);
      await rulesApi.deleteSubforumRule(
        communityType,
        subforumName,
        ruleCode
      );
      toast.success('Rule deleted successfully');
      loadRules();
    } catch (error) {
      console.error('Error deleting rule:', error);
      toast.error('Failed to delete rule');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setEditingRule(null);
    setIsCreating(false);
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
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Rules Management</CardTitle>
            <CardDescription>Configure rules for this subforum</CardDescription>
          </div>
          <Button onClick={handleCreateRule} disabled={editingRule !== null}>
            <Plus className="w-4 h-4 mr-2" />
            Add Rule
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {editingRule ? (
          <div className="space-y-4 p-4 border rounded-lg">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">
                {isCreating ? 'Create New Rule' : 'Edit Rule'}
              </h3>
              <div className="flex gap-2">
                <Button onClick={handleSaveRule} disabled={saving} size="sm">
                  <Save className="w-4 h-4 mr-2" />
                  {saving ? 'Saving...' : 'Save'}
                </Button>
                <Button onClick={handleCancel} variant="outline" size="sm">
                  <X className="w-4 h-4 mr-2" />
                  Cancel
                </Button>
              </div>
            </div>
            <Separator />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="rule-code">Rule Code</Label>
                <Input
                  id="rule-code"
                  value={editingRule.code}
                  onChange={(e) => setEditingRule({ ...editingRule, code: e.target.value })}
                  placeholder="e.g., no_harassment"
                  disabled={!isCreating}
                />
              </div>
              <div>
                <Label htmlFor="rule-name">Rule Name</Label>
                <Input
                  id="rule-name"
                  value={editingRule.name}
                  onChange={(e) => setEditingRule({ ...editingRule, name: e.target.value })}
                  placeholder="e.g., No Harassment"
                />
              </div>
            </div>
            <div>
              <Label htmlFor="rule-description">Description</Label>
              <Textarea
                id="rule-description"
                value={editingRule.description}
                onChange={(e) => setEditingRule({ ...editingRule, description: e.target.value })}
                placeholder="Describe what this rule covers and why it's important..."
                rows={3}
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="rule-category">Category</Label>
                <Select
                  value={editingRule.category}
                  onValueChange={(value) => setEditingRule({ ...editingRule, category: value })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {categories.map((category) => (
                      <SelectItem key={category.value} value={category.value}>
                        {category.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="rule-severity">Severity</Label>
                <Select
                  value={editingRule.severity}
                  onValueChange={(value) => setEditingRule({ ...editingRule, severity: value })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {severities.map((severity) => (
                      <SelectItem key={severity.value} value={severity.value}>
                        {severity.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="rule-active">Active</Label>
                <Switch
                  id="rule-active"
                  checked={editingRule.active}
                  onCheckedChange={(checked) => setEditingRule({ ...editingRule, active: checked })}
                />
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {rules.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-muted-foreground mb-4">No rules configured yet</p>
                <Button onClick={handleCreateRule}>
                  <Plus className="w-4 h-4 mr-2" />
                  Create First Rule
                </Button>
              </div>
            ) : (
              rules.map((rule) => (
                <div key={rule.code} className="border rounded-lg p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="font-semibold">{rule.name}</h3>
                        <Badge variant={rule.active ? 'default' : 'secondary'}>
                          {rule.active ? 'Active' : 'Inactive'}
                        </Badge>
                        <Badge variant="outline">{rule.category}</Badge>
                        <Badge variant="outline">{rule.severity}</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mb-2">{rule.description}</p>
                      <p className="text-xs text-muted-foreground">Code: {rule.code}</p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        onClick={() => handleEditRule(rule)}
                        variant="outline"
                        size="sm"
                        disabled={editingRule !== null}
                      >
                        <Edit className="w-4 h-4 mr-2" />
                        Edit
                      </Button>
                      <Button
                        onClick={() => handleDeleteRule(rule.code)}
                        variant="outline"
                        size="sm"
                        disabled={editingRule !== null || saving}
                      >
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                      </Button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
} 