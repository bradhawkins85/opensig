import { useState, useEffect } from 'react';
import Link from 'next/link';

type Rule = {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  template_id: string;
  conditions: Conditions;
  schedule_id?: string;
  priority: number;
  exclusive: boolean;
  active: boolean;
  created_at: string;
  updated_at: string;
};

type Conditions = {
  sender_emails?: string[];
  sender_domains?: string[];
  recipient_emails?: string[];
  recipient_domains?: string[];
  message_types?: string[];
};

type Schedule = {
  id: string;
  name: string;
};

type TestMessage = {
  sender_email: string;
  recipient_emails: string[];
  message_type: string;
  timestamp: string;
};

type EvaluationResult = {
  matched_rules: Rule[];
  selected_rule?: Rule;
};

export default function RulesPage() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [showEvaluator, setShowEvaluator] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | null>(null);
  const [tenantId, setTenantId] = useState('tenant1');
  const [evaluationResult, setEvaluationResult] = useState<EvaluationResult | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    template_id: 'template1',
    priority: 10,
    exclusive: false,
    active: true,
    schedule_id: '',
    senderEmails: '',
    senderDomains: '',
    recipientEmails: '',
    recipientDomains: '',
    messageTypes: [] as string[],
  });

  const [testMessage, setTestMessage] = useState<TestMessage>({
    sender_email: '',
    recipient_emails: [],
    message_type: 'new',
    timestamp: new Date().toISOString(),
  });

  useEffect(() => {
    fetchRules();
    fetchSchedules();
  }, [tenantId]);

  const fetchRules = async () => {
    try {
      const response = await fetch(`http://localhost:9090/v1/rules?tenant_id=${tenantId}`, {
        headers: {
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
      });
      if (response.ok) {
        const data = await response.json();
        setRules(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch rules:', error);
    }
  };

  const fetchSchedules = async () => {
    try {
      const response = await fetch(`http://localhost:9090/v1/schedules?tenant_id=${tenantId}`, {
        headers: {
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
      });
      if (response.ok) {
        const data = await response.json();
        setSchedules(data || []);
      }
    } catch (error) {
      console.error('Failed to fetch schedules:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const conditions: Conditions = {};
    if (formData.senderEmails) {
      conditions.sender_emails = formData.senderEmails.split(',').map((s) => s.trim()).filter(Boolean);
    }
    if (formData.senderDomains) {
      conditions.sender_domains = formData.senderDomains.split(',').map((s) => s.trim()).filter(Boolean);
    }
    if (formData.recipientEmails) {
      conditions.recipient_emails = formData.recipientEmails.split(',').map((s) => s.trim()).filter(Boolean);
    }
    if (formData.recipientDomains) {
      conditions.recipient_domains = formData.recipientDomains.split(',').map((s) => s.trim()).filter(Boolean);
    }
    if (formData.messageTypes.length > 0) {
      conditions.message_types = formData.messageTypes;
    }

    const ruleData = {
      tenant_id: tenantId,
      name: formData.name,
      description: formData.description,
      template_id: formData.template_id,
      conditions,
      priority: formData.priority,
      exclusive: formData.exclusive,
      active: formData.active,
      schedule_id: formData.schedule_id || undefined,
    };

    try {
      const url = editingRule
        ? `http://localhost:9090/v1/rules/${editingRule.id}`
        : 'http://localhost:9090/v1/rules';
      const method = editingRule ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
        body: JSON.stringify(ruleData),
      });

      if (response.ok) {
        fetchRules();
        resetForm();
        setShowForm(false);
      }
    } catch (error) {
      console.error('Failed to save rule:', error);
    }
  };

  const handleEdit = (rule: Rule) => {
    setEditingRule(rule);
    setFormData({
      name: rule.name,
      description: rule.description,
      template_id: rule.template_id,
      priority: rule.priority,
      exclusive: rule.exclusive,
      active: rule.active,
      schedule_id: rule.schedule_id || '',
      senderEmails: rule.conditions.sender_emails?.join(', ') || '',
      senderDomains: rule.conditions.sender_domains?.join(', ') || '',
      recipientEmails: rule.conditions.recipient_emails?.join(', ') || '',
      recipientDomains: rule.conditions.recipient_domains?.join(', ') || '',
      messageTypes: rule.conditions.message_types || [],
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this rule?')) return;

    try {
      const response = await fetch(`http://localhost:9090/v1/rules/${id}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
      });

      if (response.ok) {
        fetchRules();
      }
    } catch (error) {
      console.error('Failed to delete rule:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      template_id: 'template1',
      priority: 10,
      exclusive: false,
      active: true,
      schedule_id: '',
      senderEmails: '',
      senderDomains: '',
      recipientEmails: '',
      recipientDomains: '',
      messageTypes: [],
    });
    setEditingRule(null);
  };

  const toggleMessageType = (type: string) => {
    setFormData((prev) => ({
      ...prev,
      messageTypes: prev.messageTypes.includes(type)
        ? prev.messageTypes.filter((t) => t !== type)
        : [...prev.messageTypes, type],
    }));
  };

  const handleEvaluate = async () => {
    try {
      const response = await fetch('http://localhost:9090/v1/rules/evaluate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': 'user1',
        },
        body: JSON.stringify({
          tenant_id: tenantId,
          message: testMessage,
        }),
      });

      if (response.ok) {
        const result = await response.json();
        setEvaluationResult(result);
      }
    } catch (error) {
      console.error('Failed to evaluate rules:', error);
    }
  };

  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui' }}>
      <div style={{ marginBottom: '2rem' }}>
        <Link href="/">← Back to Home</Link>
      </div>

      <h1>Rules</h1>
      <p>Define conditions for when signatures should be applied.</p>

      <div style={{ marginBottom: '2rem', display: 'flex', gap: '1rem' }}>
        <button
          onClick={() => {
            resetForm();
            setShowForm(!showForm);
            setShowEvaluator(false);
          }}
          style={{
            padding: '0.5rem 1rem',
            background: '#0070f3',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          {showForm ? 'Cancel' : '+ New Rule'}
        </button>
        <button
          onClick={() => {
            setShowEvaluator(!showEvaluator);
            setShowForm(false);
          }}
          style={{
            padding: '0.5rem 1rem',
            background: '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          {showEvaluator ? 'Hide' : '🧪 Test Rules'}
        </button>
      </div>

      {showEvaluator && (
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #28a745', borderRadius: '4px', background: '#f0fff4' }}>
          <h2>Test Rule Evaluation</h2>
          <p>Test which rules would match for a given message.</p>
          
          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem' }}>
              Sender Email *
              <input
                type="email"
                value={testMessage.sender_email}
                onChange={(e) => setTestMessage({ ...testMessage, sender_email: e.target.value })}
                placeholder="sender@example.com"
                style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
              />
            </label>
          </div>

          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem' }}>
              Recipient Emails (comma-separated) *
              <input
                type="text"
                value={testMessage.recipient_emails.join(', ')}
                onChange={(e) => setTestMessage({ ...testMessage, recipient_emails: e.target.value.split(',').map(s => s.trim()) })}
                placeholder="recipient1@example.com, recipient2@example.com"
                style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
              />
            </label>
          </div>

          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem' }}>
              Message Type
              <select
                value={testMessage.message_type}
                onChange={(e) => setTestMessage({ ...testMessage, message_type: e.target.value })}
                style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
              >
                <option value="new">New</option>
                <option value="reply">Reply</option>
                <option value="forward">Forward</option>
              </select>
            </label>
          </div>

          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem' }}>
              Timestamp
              <input
                type="datetime-local"
                value={testMessage.timestamp.substring(0, 16)}
                onChange={(e) => setTestMessage({ ...testMessage, timestamp: new Date(e.target.value).toISOString() })}
                style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
              />
            </label>
          </div>

          <button
            onClick={handleEvaluate}
            style={{
              padding: '0.5rem 1rem',
              background: '#28a745',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Evaluate
          </button>

          {evaluationResult && (
            <div style={{ marginTop: '1rem', padding: '1rem', background: 'white', border: '1px solid #ddd', borderRadius: '4px' }}>
              <h3>Results</h3>
              <p><strong>Matched Rules:</strong> {evaluationResult.matched_rules.length}</p>
              {evaluationResult.matched_rules.length > 0 && (
                <ul>
                  {evaluationResult.matched_rules.map((rule) => (
                    <li key={rule.id}>
                      {rule.name} (Priority: {rule.priority})
                      {rule.exclusive && <span style={{ color: '#dc3545', marginLeft: '0.5rem' }}>🔒 Exclusive</span>}
                    </li>
                  ))}
                </ul>
              )}
              {evaluationResult.selected_rule && (
                <div style={{ marginTop: '1rem', padding: '0.5rem', background: '#d4edda', borderRadius: '4px' }}>
                  <strong>Selected Rule:</strong> {evaluationResult.selected_rule.name}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {showForm && (
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #ddd', borderRadius: '4px' }}>
          <h2>{editingRule ? 'Edit Rule' : 'New Rule'}</h2>
          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                Name *
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                Description
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <label style={{ display: 'block' }}>
                Template ID *
                <input
                  type="text"
                  value={formData.template_id}
                  onChange={(e) => setFormData({ ...formData, template_id: e.target.value })}
                  required
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
              <label style={{ display: 'block' }}>
                Priority
                <input
                  type="number"
                  value={formData.priority}
                  onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block' }}>
                Schedule (optional)
                <select
                  value={formData.schedule_id}
                  onChange={(e) => setFormData({ ...formData, schedule_id: e.target.value })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                >
                  <option value="">No schedule</option>
                  {schedules.map((schedule) => (
                    <option key={schedule.id} value={schedule.id}>
                      {schedule.name}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <fieldset style={{ marginBottom: '1rem', padding: '1rem', border: '1px solid #ddd', borderRadius: '4px' }}>
              <legend>Conditions</legend>
              
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                  Sender Emails (comma-separated)
                  <input
                    type="text"
                    value={formData.senderEmails}
                    onChange={(e) => setFormData({ ...formData, senderEmails: e.target.value })}
                    placeholder="user1@example.com, user2@example.com"
                    style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                  />
                </label>
              </div>

              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                  Sender Domains (comma-separated)
                  <input
                    type="text"
                    value={formData.senderDomains}
                    onChange={(e) => setFormData({ ...formData, senderDomains: e.target.value })}
                    placeholder="example.com, company.com"
                    style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                  />
                </label>
              </div>

              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                  Recipient Emails (comma-separated)
                  <input
                    type="text"
                    value={formData.recipientEmails}
                    onChange={(e) => setFormData({ ...formData, recipientEmails: e.target.value })}
                    placeholder="client@external.com, partner@company.com"
                    style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                  />
                </label>
              </div>

              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>
                  Recipient Domains (comma-separated)
                  <input
                    type="text"
                    value={formData.recipientDomains}
                    onChange={(e) => setFormData({ ...formData, recipientDomains: e.target.value })}
                    placeholder="client.com, external.com"
                    style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                  />
                </label>
              </div>

              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>Message Types:</label>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  {['new', 'reply', 'forward'].map((type) => (
                    <button
                      key={type}
                      type="button"
                      onClick={() => toggleMessageType(type)}
                      style={{
                        padding: '0.5rem 1rem',
                        background: formData.messageTypes.includes(type) ? '#0070f3' : '#e0e0e0',
                        color: formData.messageTypes.includes(type) ? 'white' : 'black',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        textTransform: 'capitalize',
                      }}
                    >
                      {type}
                    </button>
                  ))}
                </div>
              </div>
            </fieldset>

            <div style={{ marginBottom: '1rem', display: 'flex', gap: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.exclusive}
                  onChange={(e) => setFormData({ ...formData, exclusive: e.target.checked })}
                />
                Exclusive (stop processing other rules if this matches)
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.active}
                  onChange={(e) => setFormData({ ...formData, active: e.target.checked })}
                />
                Active
              </label>
            </div>

            <button
              type="submit"
              style={{
                padding: '0.5rem 1rem',
                background: '#0070f3',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                marginRight: '0.5rem',
              }}
            >
              {editingRule ? 'Update' : 'Create'} Rule
            </button>
            <button
              type="button"
              onClick={() => {
                resetForm();
                setShowForm(false);
              }}
              style={{
                padding: '0.5rem 1rem',
                background: '#ccc',
                color: 'black',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
          </form>
        </div>
      )}

      <div>
        <h2>Existing Rules</h2>
        {rules.length === 0 ? (
          <p style={{ color: '#666' }}>No rules yet. Create one to get started.</p>
        ) : (
          <div style={{ display: 'grid', gap: '1rem' }}>
            {rules.map((rule) => (
              <div key={rule.id} style={{ padding: '1rem', border: '1px solid #ddd', borderRadius: '4px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                  <div style={{ flex: 1 }}>
                    <h3 style={{ margin: '0 0 0.5rem 0' }}>
                      {rule.name}
                      {!rule.active && <span style={{ color: '#999', marginLeft: '0.5rem' }}>(Inactive)</span>}
                      {rule.exclusive && <span style={{ color: '#dc3545', marginLeft: '0.5rem' }}>🔒 Exclusive</span>}
                    </h3>
                    {rule.description && <p style={{ margin: '0 0 0.5rem 0', color: '#666' }}>{rule.description}</p>}
                    <div style={{ fontSize: '0.9rem', color: '#333', marginTop: '0.5rem' }}>
                      <div><strong>Template:</strong> {rule.template_id}</div>
                      <div><strong>Priority:</strong> {rule.priority}</div>
                      {rule.schedule_id && (
                        <div><strong>Schedule:</strong> {schedules.find((s) => s.id === rule.schedule_id)?.name || rule.schedule_id}</div>
                      )}
                      <div style={{ marginTop: '0.5rem' }}>
                        <strong>Conditions:</strong>
                        <ul style={{ margin: '0.25rem 0 0 1.5rem', padding: 0 }}>
                          {rule.conditions.sender_emails && rule.conditions.sender_emails.length > 0 && (
                            <li>Sender emails: {rule.conditions.sender_emails.join(', ')}</li>
                          )}
                          {rule.conditions.sender_domains && rule.conditions.sender_domains.length > 0 && (
                            <li>Sender domains: {rule.conditions.sender_domains.join(', ')}</li>
                          )}
                          {rule.conditions.recipient_emails && rule.conditions.recipient_emails.length > 0 && (
                            <li>Recipient emails: {rule.conditions.recipient_emails.join(', ')}</li>
                          )}
                          {rule.conditions.recipient_domains && rule.conditions.recipient_domains.length > 0 && (
                            <li>Recipient domains: {rule.conditions.recipient_domains.join(', ')}</li>
                          )}
                          {rule.conditions.message_types && rule.conditions.message_types.length > 0 && (
                            <li>Message types: {rule.conditions.message_types.join(', ')}</li>
                          )}
                        </ul>
                      </div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <button
                      onClick={() => handleEdit(rule)}
                      style={{
                        padding: '0.25rem 0.75rem',
                        background: '#0070f3',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                      }}
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(rule.id)}
                      style={{
                        padding: '0.25rem 0.75rem',
                        background: '#dc3545',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
