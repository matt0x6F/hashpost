'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Textarea } from '@/components/shadcn/textarea';
import { Switch } from '@/components/shadcn/switch';
import { Badge } from '@/components/shadcn/badge';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/shadcn/select';
import { 
  Plus, 
  Trash2, 
  Save,
  Loader2
} from 'lucide-react';
import { toast } from 'sonner';

interface SubforumRulesManagerProps {
  communityType: string;
  subforumName: string;
  subforumId: string;
}

export function SubforumRulesManager({ communityType, subforumName, subforumId }: SubforumRulesManagerProps) {
  const [rules, setRules] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadRules();
  }, [subforumId]);

  const loadRules = async () => {
    try {
      setLoading(true);
      // Subforum rules not available in atproto system
      setRules([]);
    } catch (error) {
      console.error('Error loading rules:', error);
      toast.error('Failed to load rules');
    } finally {
      setLoading(false);
    }
  };

  const saveRules = async (rulesToSave: any[]) => {
    try {
      setSaving(true);
      
      // Subforum rules not available in atproto system
      toast.error('Subforum rules are not available in the atproto system');
    } catch (error) {
      console.error('Error saving rules:', error);
      toast.error('Failed to save rules');
    } finally {
      setSaving(false);
    }
  };

  const addRule = () => {
    const newRule = {
      id: Date.now(),
      code: '',
      name: '',
      description: '',
      category: 'content',
      severity: 'medium',
      active: true
    };
    setRules([...rules, newRule]);
  };

  const updateRule = (id: number, field: string, value: any) => {
    setRules(rules.map(rule => 
      rule.id === id ? { ...rule, [field]: value } : rule
    ));
  };

  const removeRule = (id: number) => {
    setRules(rules.filter(rule => rule.id !== id));
  };

  const handleSave = () => {
    saveRules(rules);
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            Loading rules...
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Subforum Rules</CardTitle>
          <p className="text-sm text-muted-foreground">
            Subforum rules are not available in the atproto system
          </p>
        </CardHeader>
        <CardContent>
          <div className="text-center text-muted-foreground py-8">
            <p>Subforum rules management is not available in the atproto system.</p>
            <p className="text-sm mt-2">
              In the atproto system, content moderation is handled through different mechanisms.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}