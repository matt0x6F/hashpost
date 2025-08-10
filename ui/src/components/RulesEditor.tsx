'use client';

import { useState } from 'react';
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

export interface Rule {
  code: string;
  name: string;
  description: string;
  category: string;
  severity: string;
  active: boolean;
}

interface RulesEditorProps {
  rules: Rule[];
  onChange: (rules: Rule[]) => void;
  disabled?: boolean;
  showTitle?: boolean;
  maxRules?: number;
}

export function RulesEditor({ 
  rules, 
  onChange, 
  disabled = false, 
  showTitle = true,
  maxRules = 10 
}: RulesEditorProps) {
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

  const handleCreateRule = () => {
    if (rules.length >= maxRules) {
      return;
    }
    
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

  const handleSaveRule = () => {
    if (!editingRule) return;

    // Generate code from name if empty or if creating
    if (isCreating && !editingRule.code.trim()) {
      const generatedCode = editingRule.name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '_')
        .replace(/^_+|_+$/g, '');
      editingRule.code = generatedCode;
    }

    // Validate required fields
    if (!editingRule.code.trim() || !editingRule.name.trim() || !editingRule.description.trim()) {
      return;
    }

    // Check for duplicate codes
    const existingRule = isCreating 
      ? rules.find(r => r.code === editingRule.code)
      : rules.find(r => r.code === editingRule.code && r !== editingRule);
    if (existingRule) {
      return;
    }

    let updatedRules: Rule[];
    if (isCreating) {
      updatedRules = [...rules, editingRule];
    } else {
      updatedRules = rules.map(r => r.code === editingRule.code ? editingRule : r);
    }

    onChange(updatedRules);
    setEditingRule(null);
    setIsCreating(false);
  };

  const handleDeleteRule = (ruleCode: string) => {
    const updatedRules = rules.filter(r => r.code !== ruleCode);
    onChange(updatedRules);
  };

  const handleCancel = () => {
    setEditingRule(null);
    setIsCreating(false);
  };

  return (
    <Card>
      {showTitle && (
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Community Rules</CardTitle>
              <CardDescription>Configure rules and guidelines for this community</CardDescription>
            </div>
            <Button 
              onClick={handleCreateRule} 
              disabled={disabled || editingRule !== null || rules.length >= maxRules}
              size="sm"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Rule
            </Button>
          </div>
        </CardHeader>
      )}
      <CardContent className="space-y-6">
        {editingRule ? (
          <div className="space-y-4 p-4 border rounded-lg">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">
                {isCreating ? 'Create New Rule' : 'Edit Rule'}
              </h3>
              <div className="flex gap-2">
                <Button onClick={handleSaveRule} disabled={disabled} size="sm">
                  <Save className="w-4 h-4 mr-2" />
                  Save
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
                  disabled={!isCreating || disabled}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Unique identifier for this rule
                </p>
              </div>
              <div>
                <Label htmlFor="rule-name">Rule Name</Label>
                <Input
                  id="rule-name"
                  value={editingRule.name}
                  onChange={(e) => setEditingRule({ ...editingRule, name: e.target.value })}
                  placeholder="e.g., No Harassment"
                  disabled={disabled}
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
                disabled={disabled}
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="rule-category">Category</Label>
                <Select
                  value={editingRule.category}
                  onValueChange={(value) => setEditingRule({ ...editingRule, category: value })}
                  disabled={disabled}
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
                  disabled={disabled}
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
                  disabled={disabled}
                />
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {!showTitle && rules.length < maxRules && (
              <div className="flex justify-between items-center">
                <Label className="text-sm font-medium">Community Rules</Label>
                <Button 
                  onClick={handleCreateRule} 
                  disabled={disabled || editingRule !== null}
                  size="sm"
                  variant="outline"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Add Rule
                </Button>
              </div>
            )}
            
            {rules.length === 0 ? (
              <div className="text-center py-6 border-2 border-dashed border-muted rounded-lg">
                <p className="text-muted-foreground mb-4">No rules configured</p>
                {!disabled && (
                  <Button onClick={handleCreateRule} variant="outline" size="sm">
                    <Plus className="w-4 h-4 mr-2" />
                    Create First Rule
                  </Button>
                )}
              </div>
            ) : (
              <div className="space-y-3">
                {rules.map((rule) => (
                  <div key={rule.code} className="border rounded-lg p-3">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <h4 className="font-medium">{rule.name}</h4>
                          <Badge variant={rule.active ? 'default' : 'secondary'} className="text-xs">
                            {rule.active ? 'Active' : 'Inactive'}
                          </Badge>
                          <Badge variant="outline" className="text-xs">{rule.category}</Badge>
                          <Badge variant="outline" className="text-xs">{rule.severity}</Badge>
                        </div>
                        <p className="text-sm text-muted-foreground mb-1">{rule.description}</p>
                        <p className="text-xs text-muted-foreground">Code: {rule.code}</p>
                      </div>
                      {!disabled && (
                        <div className="flex gap-1">
                          <Button
                            onClick={() => handleEditRule(rule)}
                            variant="ghost"
                            size="sm"
                            disabled={editingRule !== null}
                          >
                            <Edit className="w-3 h-3" />
                          </Button>
                          <Button
                            onClick={() => handleDeleteRule(rule.code)}
                            variant="ghost"
                            size="sm"
                            disabled={editingRule !== null}
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
            
            {rules.length >= maxRules && (
              <p className="text-xs text-muted-foreground text-center">
                Maximum of {maxRules} rules allowed
              </p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
