"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Textarea } from "@/components/shadcn/textarea";
import { Badge } from "@/components/shadcn/badge";
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/shadcn/select";
import { 
  Search, 
  CheckCircle, 
  Clock,
  FileText,
  Gavel
} from "lucide-react";
import { toast } from "sonner";

interface CorrelationRequest {
  id: string;
  requestedPseudonym: string;
  justification: string;
  legalBasis: string;
  incidentId: string;
  scope: string;
  status: "pending" | "approved" | "rejected" | "completed";
  requestedBy: string;
  createdAt: string;
  completedAt?: string;
  result?: string;
}

const mockRequests: CorrelationRequest[] = [
  {
    id: "1",
    requestedPseudonym: "user123",
    justification: "Investigation of coordinated harassment campaign",
    legalBasis: "Platform safety investigation",
    incidentId: "INC-2024-001",
    scope: "platform_wide",
    status: "completed",
    requestedBy: "admin1",
    createdAt: "2024-01-10T09:00:00Z",
    completedAt: "2024-01-10T14:30:00Z",
    result: "3 pseudonyms identified, action taken"
  },
  {
    id: "2",
    requestedPseudonym: "user456",
    justification: "Legal compliance request from law enforcement",
    legalBasis: "Court order",
    incidentId: "LEGAL-2024-002",
    scope: "legal_compliance",
    status: "approved",
    requestedBy: "admin2",
    createdAt: "2024-01-12T11:00:00Z"
  },
  {
    id: "3",
    requestedPseudonym: "user789",
    justification: "Trust and safety investigation",
    legalBasis: "Platform safety investigation",
    incidentId: "TS-2024-003",
    scope: "trust_safety",
    status: "pending",
    requestedBy: "admin3",
    createdAt: "2024-01-15T16:00:00Z"
  }
];

export function CorrelationTab() {
  const [requests, setRequests] = useState<CorrelationRequest[]>(mockRequests);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [scopeFilter, setScopeFilter] = useState<string>("all");
  const [newRequest, setNewRequest] = useState({
    requestedPseudonym: "",
    justification: "",
    legalBasis: "",
    incidentId: "",
    scope: "platform_wide"
  });

  // Filter requests based on selected filters
  const filteredRequests = requests.filter(request => {
    if (statusFilter !== "all" && request.status !== statusFilter) return false;
    if (scopeFilter !== "all" && request.scope !== scopeFilter) return false;
    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return <Badge variant="secondary">Pending</Badge>;
      case "approved":
        return <Badge variant="default">Approved</Badge>;
      case "rejected":
        return <Badge variant="destructive">Rejected</Badge>;
      case "completed":
        return <Badge variant="outline">Completed</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  const getScopeBadge = (scope: string) => {
    switch (scope) {
      case "platform_wide":
        return <Badge variant="default">Platform Wide</Badge>;
      case "legal_compliance":
        return <Badge variant="destructive">Legal Compliance</Badge>;
      case "trust_safety":
        return <Badge variant="secondary">Trust & Safety</Badge>;
      default:
        return <Badge variant="outline">{scope}</Badge>;
    }
  };

  const handleSubmitRequest = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!newRequest.requestedPseudonym || !newRequest.justification) {
      toast.error("Please fill in all required fields");
      return;
    }

    const request: CorrelationRequest = {
      id: Date.now().toString(),
      requestedPseudonym: newRequest.requestedPseudonym,
      justification: newRequest.justification,
      legalBasis: newRequest.legalBasis,
      incidentId: newRequest.incidentId,
      scope: newRequest.scope,
      status: "pending",
      requestedBy: "current_admin",
      createdAt: new Date().toISOString()
    };

    setRequests([request, ...requests]);
    setNewRequest({
      requestedPseudonym: "",
      justification: "",
      legalBasis: "",
      incidentId: "",
      scope: "platform_wide"
    });
    
    toast.success("Correlation request submitted");
  };

  const handleAction = (requestId: string, action: "pending" | "approved" | "rejected" | "completed") => {
    setRequests(requests.map(req => 
      req.id === requestId 
        ? { ...req, status: action, completedAt: new Date().toISOString() }
        : req
    ));
    toast.success(`Request ${action}`);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Statistics Overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pending Requests</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {requests.filter(r => r.status === "pending").length}
            </div>
            <p className="text-xs text-muted-foreground">
              Awaiting approval
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Completed Today</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {requests.filter(r => r.status === "completed" && 
                new Date(r.completedAt!).toDateString() === new Date().toDateString()).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Successfully processed
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Legal Requests</CardTitle>
            <Gavel className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {requests.filter(r => r.scope === "legal_compliance").length}
            </div>
            <p className="text-xs text-muted-foreground">
              Legal compliance
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{requests.length}</div>
            <p className="text-xs text-muted-foreground">
              This period
            </p>
          </CardContent>
        </Card>
      </div>

      {/* New Request Form */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            New Correlation Request
          </CardTitle>
          <CardDescription>
            Request cross-user identity correlation for platform-wide investigations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmitRequest} className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Target Pseudonym</label>
                <Input
                  placeholder="Enter pseudonym ID"
                  value={newRequest.requestedPseudonym}
                  onChange={(e) => setNewRequest({...newRequest, requestedPseudonym: e.target.value})}
                  required
                />
              </div>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Scope</label>
                <Select 
                  value={newRequest.scope} 
                  onValueChange={(value) => setNewRequest({...newRequest, scope: value})}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="platform_wide">Platform Wide</SelectItem>
                    <SelectItem value="legal_compliance">Legal Compliance</SelectItem>
                    <SelectItem value="trust_safety">Trust & Safety</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Justification</label>
              <Textarea
                placeholder="Explain why this correlation is needed..."
                value={newRequest.justification}
                onChange={(e) => setNewRequest({...newRequest, justification: e.target.value})}
                required
                rows={3}
              />
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Legal Basis</label>
                <Input
                  placeholder="Legal basis for request"
                  value={newRequest.legalBasis}
                  onChange={(e) => setNewRequest({...newRequest, legalBasis: e.target.value})}
                />
              </div>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Incident ID</label>
                <Input
                  placeholder="Optional incident reference"
                  value={newRequest.incidentId}
                  onChange={(e) => setNewRequest({...newRequest, incidentId: e.target.value})}
                />
              </div>
            </div>

            <Button type="submit" className="w-full">
              Submit Correlation Request
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Status</label>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="approved">Approved</SelectItem>
                  <SelectItem value="rejected">Rejected</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Scope</label>
              <Select value={scopeFilter} onValueChange={setScopeFilter}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="All Scopes" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Scopes</SelectItem>
                  <SelectItem value="platform_wide">Platform Wide</SelectItem>
                  <SelectItem value="legal_compliance">Legal Compliance</SelectItem>
                  <SelectItem value="trust_safety">Trust & Safety</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Requests Table */}
      <Card>
        <CardHeader>
          <CardTitle>Correlation Requests</CardTitle>
          <CardDescription>
            Track all identity correlation requests and their status
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {filteredRequests.map((request) => (
              <div key={request.id} className="border rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Badge variant="outline">{request.requestedPseudonym}</Badge>
                    {getStatusBadge(request.status)}
                    {getScopeBadge(request.scope)}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {formatDate(request.createdAt)}
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div>
                    <span className="font-medium">Justification:</span> {request.justification}
                  </div>
                  {request.legalBasis && (
                    <div>
                      <span className="font-medium">Legal Basis:</span> {request.legalBasis}
                    </div>
                  )}
                  {request.incidentId && (
                    <div>
                      <span className="font-medium">Incident ID:</span> {request.incidentId}
                    </div>
                  )}
                  {request.result && (
                    <div>
                      <span className="font-medium">Result:</span> {request.result}
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    Requested by: {request.requestedBy}
                  </div>
                  
                  {request.status === "pending" && (
                    <div className="flex gap-2">
                      <Button 
                        size="sm" 
                        onClick={() => handleAction(request.id, "approved")}
                      >
                        Approve
                      </Button>
                      <Button 
                        size="sm" 
                        variant="destructive"
                        onClick={() => handleAction(request.id, "rejected")}
                      >
                        Reject
                      </Button>
                    </div>
                  )}
                  
                  {request.status === "approved" && (
                    <Button 
                      size="sm"
                      onClick={() => handleAction(request.id, "completed")}
                    >
                      Mark Complete
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
