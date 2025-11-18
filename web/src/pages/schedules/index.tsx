import { useState, useEffect } from 'react';
import Link from 'next/link';

type Schedule = {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  active: boolean;
  time_windows: TimeWindow[];
  created_at: string;
  updated_at: string;
};

type TimeWindow = {
  start_date: string;
  end_date: string;
  start_time?: string;
  end_time?: string;
  timezone: string;
  recurrence?: Recurrence;
};

type Recurrence = {
  frequency: 'daily' | 'weekly' | 'monthly' | 'yearly';
  interval: number;
  days_of_week?: number[];
  day_of_month?: number;
  end_date?: string;
};

export default function SchedulesPage() {
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<Schedule | null>(null);
  const [tenantId, setTenantId] = useState('tenant1'); // Default tenant for demo

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    active: true,
    startDate: '',
    endDate: '',
    startTime: '',
    endTime: '',
    timezone: 'UTC',
    hasRecurrence: false,
    frequency: 'weekly' as 'daily' | 'weekly' | 'monthly' | 'yearly',
    interval: 1,
    daysOfWeek: [] as number[],
  });

  useEffect(() => {
    fetchSchedules();
  }, [tenantId]);

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

    const timeWindow: TimeWindow = {
      start_date: new Date(formData.startDate).toISOString(),
      end_date: new Date(formData.endDate).toISOString(),
      timezone: formData.timezone,
    };

    if (formData.startTime) {
      timeWindow.start_time = formData.startTime;
    }
    if (formData.endTime) {
      timeWindow.end_time = formData.endTime;
    }

    if (formData.hasRecurrence) {
      timeWindow.recurrence = {
        frequency: formData.frequency,
        interval: formData.interval,
      };
      if (formData.frequency === 'weekly' && formData.daysOfWeek.length > 0) {
        timeWindow.recurrence.days_of_week = formData.daysOfWeek;
      }
    }

    const scheduleData = {
      tenant_id: tenantId,
      name: formData.name,
      description: formData.description,
      active: formData.active,
      time_windows: [timeWindow],
    };

    try {
      const url = editingSchedule
        ? `http://localhost:9090/v1/schedules/${editingSchedule.id}`
        : 'http://localhost:9090/v1/schedules';
      const method = editingSchedule ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
        body: JSON.stringify(scheduleData),
      });

      if (response.ok) {
        fetchSchedules();
        resetForm();
        setShowForm(false);
      }
    } catch (error) {
      console.error('Failed to save schedule:', error);
    }
  };

  const handleEdit = (schedule: Schedule) => {
    setEditingSchedule(schedule);
    const window = schedule.time_windows[0] || {};
    setFormData({
      name: schedule.name,
      description: schedule.description,
      active: schedule.active,
      startDate: window.start_date ? window.start_date.split('T')[0] : '',
      endDate: window.end_date ? window.end_date.split('T')[0] : '',
      startTime: window.start_time || '',
      endTime: window.end_time || '',
      timezone: window.timezone || 'UTC',
      hasRecurrence: !!window.recurrence,
      frequency: window.recurrence?.frequency || 'weekly',
      interval: window.recurrence?.interval || 1,
      daysOfWeek: window.recurrence?.days_of_week || [],
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this schedule?')) return;

    try {
      const response = await fetch(`http://localhost:9090/v1/schedules/${id}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': 'user1',
          'X-User-Role': 'signature_admin',
        },
      });

      if (response.ok) {
        fetchSchedules();
      }
    } catch (error) {
      console.error('Failed to delete schedule:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      active: true,
      startDate: '',
      endDate: '',
      startTime: '',
      endTime: '',
      timezone: 'UTC',
      hasRecurrence: false,
      frequency: 'weekly',
      interval: 1,
      daysOfWeek: [],
    });
    setEditingSchedule(null);
  };

  const toggleDayOfWeek = (day: number) => {
    setFormData((prev) => ({
      ...prev,
      daysOfWeek: prev.daysOfWeek.includes(day)
        ? prev.daysOfWeek.filter((d) => d !== day)
        : [...prev.daysOfWeek, day],
    }));
  };

  const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui' }}>
      <div style={{ marginBottom: '2rem' }}>
        <Link href="/">← Back to Home</Link>
      </div>

      <h1>Schedules</h1>
      <p>Manage time windows and recurrence patterns for signature rules.</p>

      <div style={{ marginBottom: '2rem' }}>
        <button
          onClick={() => {
            resetForm();
            setShowForm(!showForm);
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
          {showForm ? 'Cancel' : '+ New Schedule'}
        </button>
      </div>

      {showForm && (
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #ddd', borderRadius: '4px' }}>
          <h2>{editingSchedule ? 'Edit Schedule' : 'New Schedule'}</h2>
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
                Start Date *
                <input
                  type="date"
                  value={formData.startDate}
                  onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                  required
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
              <label style={{ display: 'block' }}>
                End Date *
                <input
                  type="date"
                  value={formData.endDate}
                  onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                  required
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <label style={{ display: 'block' }}>
                Start Time (HH:MM)
                <input
                  type="time"
                  value={formData.startTime}
                  onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
              <label style={{ display: 'block' }}>
                End Time (HH:MM)
                <input
                  type="time"
                  value={formData.endTime}
                  onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                />
              </label>
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block' }}>
                Timezone
                <select
                  value={formData.timezone}
                  onChange={(e) => setFormData({ ...formData, timezone: e.target.value })}
                  style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                >
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">America/New_York</option>
                  <option value="America/Chicago">America/Chicago</option>
                  <option value="America/Los_Angeles">America/Los_Angeles</option>
                  <option value="Europe/London">Europe/London</option>
                </select>
              </label>
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.hasRecurrence}
                  onChange={(e) => setFormData({ ...formData, hasRecurrence: e.target.checked })}
                />
                Enable Recurrence
              </label>
            </div>

            {formData.hasRecurrence && (
              <div style={{ marginLeft: '1.5rem', marginBottom: '1rem', padding: '1rem', background: '#f5f5f5', borderRadius: '4px' }}>
                <div style={{ marginBottom: '1rem' }}>
                  <label style={{ display: 'block' }}>
                    Frequency
                    <select
                      value={formData.frequency}
                      onChange={(e) => setFormData({ ...formData, frequency: e.target.value as any })}
                      style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                    >
                      <option value="daily">Daily</option>
                      <option value="weekly">Weekly</option>
                      <option value="monthly">Monthly</option>
                      <option value="yearly">Yearly</option>
                    </select>
                  </label>
                </div>

                <div style={{ marginBottom: '1rem' }}>
                  <label style={{ display: 'block' }}>
                    Repeat every
                    <input
                      type="number"
                      min="1"
                      value={formData.interval}
                      onChange={(e) => setFormData({ ...formData, interval: parseInt(e.target.value) })}
                      style={{ display: 'block', width: '100%', padding: '0.5rem', marginTop: '0.25rem' }}
                    />
                  </label>
                </div>

                {formData.frequency === 'weekly' && (
                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem' }}>Days of Week:</label>
                    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                      {dayNames.map((day, index) => (
                        <button
                          key={index}
                          type="button"
                          onClick={() => toggleDayOfWeek(index)}
                          style={{
                            padding: '0.5rem 1rem',
                            background: formData.daysOfWeek.includes(index) ? '#0070f3' : '#e0e0e0',
                            color: formData.daysOfWeek.includes(index) ? 'white' : 'black',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                          }}
                        >
                          {day.substring(0, 3)}
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            <div style={{ marginBottom: '1rem' }}>
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
              {editingSchedule ? 'Update' : 'Create'} Schedule
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
        <h2>Existing Schedules</h2>
        {schedules.length === 0 ? (
          <p style={{ color: '#666' }}>No schedules yet. Create one to get started.</p>
        ) : (
          <div style={{ display: 'grid', gap: '1rem' }}>
            {schedules.map((schedule) => (
              <div key={schedule.id} style={{ padding: '1rem', border: '1px solid #ddd', borderRadius: '4px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                  <div>
                    <h3 style={{ margin: '0 0 0.5rem 0' }}>
                      {schedule.name}
                      {!schedule.active && <span style={{ color: '#999', marginLeft: '0.5rem' }}>(Inactive)</span>}
                    </h3>
                    {schedule.description && <p style={{ margin: '0 0 0.5rem 0', color: '#666' }}>{schedule.description}</p>}
                    {schedule.time_windows.map((window, idx) => (
                      <div key={idx} style={{ fontSize: '0.9rem', color: '#333' }}>
                        <div>
                          📅 {new Date(window.start_date).toLocaleDateString()} - {new Date(window.end_date).toLocaleDateString()}
                        </div>
                        {window.start_time && window.end_time && (
                          <div>⏰ {window.start_time} - {window.end_time} ({window.timezone})</div>
                        )}
                        {window.recurrence && (
                          <div>
                            🔄 Repeats {window.recurrence.frequency} (every {window.recurrence.interval})
                            {window.recurrence.days_of_week && (
                              <> on {window.recurrence.days_of_week.map((d) => dayNames[d].substring(0, 3)).join(', ')}</>
                            )}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                    <button
                      onClick={() => handleEdit(schedule)}
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
                      onClick={() => handleDelete(schedule.id)}
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
