import { useState, useEffect, useRef } from 'react';
import billyIcon from './assets/images/billy-icon.png';
import './App.css';
import {
  GetStatus, SendMessage, ListModels, PopOut, GetPlatform,
  GetConversations, GetMessages, GetMemories, AddMemory, DeleteMemory,
  SetActiveConversation, NewConversation,
} from '../wailsjs/go/main/App';
import { Events } from '@wailsio/runtime';
import { main } from '../wailsjs/go/models';

type Theme = 'system' | 'dark' | 'light';
type SidebarTab = 'history' | 'memories';

function applyTheme(theme: Theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('billy-theme', theme);
}

function App() {
  const [messages, setMessages] = useState<main.Message[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [status, setStatus] = useState<main.StatusInfo | null>(null);
  const [models, setModels] = useState<string[]>([]);
  const [activeModel, setActiveModel] = useState('qwen2.5-coder:7b');
  const [activeConvID, setActiveConvID] = useState('');
  const [activeConvTitle, setActiveConvTitle] = useState('');
  const [showModelPicker, setShowModelPicker] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showSidebar, setShowSidebar] = useState(false);
  const [sidebarTab, setSidebarTab] = useState<SidebarTab>('history');
  const [conversations, setConversations] = useState<main.ConversationSummary[]>([]);
  const [memories, setMemories] = useState<main.MemoryItem[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [newMemory, setNewMemory] = useState('');
  const [theme, setTheme] = useState<Theme>(() => {
    return (localStorage.getItem('billy-theme') as Theme) || 'system';
  });
  const [platform, setPlatform] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);
  const streamBuffer = useRef('');

  useEffect(() => { applyTheme(theme); }, [theme]);

  function changeTheme(t: Theme) {
    setTheme(t);
    applyTheme(t);
    setShowSettings(false);
  }

  useEffect(() => {
    GetStatus().then(s => {
      setStatus(s);
      if (s.activeModel) setActiveModel(s.activeModel);
    });
    ListModels().then(m => { if (m) setModels(m); });
    GetPlatform().then(setPlatform);

    Events.On('chat:token', (ev) => {
      const token = ev.data as string;
      streamBuffer.current += token;
      setMessages(prev => {
        const msgs = [...prev];
        if (msgs.length > 0 && msgs[msgs.length - 1].role === 'assistant') {
          msgs[msgs.length - 1] = { ...msgs[msgs.length - 1], content: streamBuffer.current };
        }
        return msgs;
      });
    });

    Events.On('chat:done', (ev) => {
      const convID = ev.data as string;
      streamBuffer.current = '';
      setStreaming(false);
      if (convID) setActiveConvID(convID);
    });

    Events.On('chat:error', (ev) => {
      const err = ev.data as string;
      streamBuffer.current = '';
      setStreaming(false);
      setMessages(prev => [...prev, new main.Message({ role: 'assistant', content: `⚠️ ${err}` })]);
    });

    // AI finished generating a title — update conversation list and header
    Events.On('conv:titled', (ev) => {
      const data = ev.data as { id: string; title: string };
      setActiveConvTitle(data.title);
      setConversations(prev =>
        prev.map(c => c.id === data.id ? { ...c, title: data.title } : c)
      );
    });

    const interval = setInterval(() => {
      GetStatus().then(setStatus);
    }, 15000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  function sendMessage() {
    const text = input.trim();
    if (!text || streaming) return;
    setInput('');

    const newMessages: main.Message[] = [
      ...messages,
      new main.Message({ role: 'user', content: text }),
    ];
    setMessages([...newMessages, new main.Message({ role: 'assistant', content: '' })]);
    setStreaming(true);
    streamBuffer.current = '';

    SendMessage(new main.ChatRequest({ messages: newMessages, model: activeModel, userText: text }));
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  function clearChat() {
    setMessages([]);
    setActiveConvID('');
    setActiveConvTitle('');
    streamBuffer.current = '';
    NewConversation();
  }

  async function openSidebar(tab: SidebarTab) {
    setSidebarTab(tab);
    setShowSidebar(true);
    setLoadingHistory(true);
    if (tab === 'history') {
      const convs = await GetConversations();
      setConversations(convs || []);
    } else {
      const mems = await GetMemories();
      setMemories(mems || []);
    }
    setLoadingHistory(false);
  }

  async function loadConversation(conv: main.ConversationSummary) {
    const msgs = await GetMessages(conv.id);
    if (!msgs || msgs.length === 0) return;
    setMessages(msgs.map(m => new main.Message({ role: m.role, content: m.content })));
    setActiveConvID(conv.id);
    setActiveConvTitle(conv.title);
    SetActiveConversation(conv.id);
    setShowSidebar(false);
  }

  async function saveMemory() {
    const text = newMemory.trim();
    if (!text) return;
    const id = await AddMemory(text);
    if (id) {
      const now = new Date();
      const dateStr = now.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
      setMemories(prev => [...prev, new main.MemoryItem({ id, content: text, createdAt: dateStr })]);
      setNewMemory('');
    }
  }

  async function removeMemory(id: string) {
    await DeleteMemory(id);
    setMemories(prev => prev.filter(m => m.id !== id));
  }

  const ollamaOffline = status && !status.ollamaReady;

  return (
    <div className="app">
      {/* Title bar */}
      <div className="titlebar" style={{ '--wails-draggable': 'drag' } as React.CSSProperties}>
        <div className="titlebar-left">
          <img src={billyIcon} className="titlebar-icon" alt="Billy" />
          <span className="titlebar-name">
            {activeConvTitle || 'Billy'}
          </span>
          {status && (
            <span className="tier-badge">{status.tier}</span>
          )}
        </div>
        <div className="titlebar-right">
          <button className="icon-btn" title="History" onClick={() => showSidebar && sidebarTab === 'history' ? setShowSidebar(false) : openSidebar('history')}>🕐</button>
          <button className="icon-btn" title="Memories" onClick={() => showSidebar && sidebarTab === 'memories' ? setShowSidebar(false) : openSidebar('memories')}>🧠</button>
          <button className="icon-btn" title="Settings" onClick={() => setShowSettings(v => !v)}>⚙</button>
          <button className="icon-btn" title="Pop out" onClick={() => PopOut()}>⤢</button>
          <button className="icon-btn" title="New chat" onClick={clearChat}>⊘</button>
        </div>
      </div>

      {/* Settings panel */}
      {showSettings && (
        <>
          <div className="settings-overlay" onClick={() => setShowSettings(false)} />
          <div className="settings-panel">
            <div className="settings-title">Appearance</div>
            <div className="theme-options">
              {([
                { value: 'system', icon: '💻', label: 'System' },
                { value: 'dark',   icon: '🌙', label: 'Dark'   },
                { value: 'light',  icon: '☀️',  label: 'Light'  },
              ] as { value: Theme; icon: string; label: string }[]).map(opt => (
                <button
                  key={opt.value}
                  className={`theme-btn${theme === opt.value ? ' active' : ''}`}
                  onClick={() => changeTheme(opt.value)}
                >
                  <span className="theme-icon">{opt.icon}</span>
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
        </>
      )}

      {/* Sidebar: history + memories */}
      {showSidebar && (
        <>
          <div className="sidebar-overlay" onClick={() => setShowSidebar(false)} />
          <div className="sidebar">
            <div className="sidebar-tabs">
              <button className={`sidebar-tab${sidebarTab === 'history' ? ' active' : ''}`} onClick={() => openSidebar('history')}>History</button>
              <button className={`sidebar-tab${sidebarTab === 'memories' ? ' active' : ''}`} onClick={() => openSidebar('memories')}>Memories</button>
            </div>
            <div className="sidebar-content">
              {loadingHistory && <div className="sidebar-empty">Loading…</div>}

              {!loadingHistory && sidebarTab === 'history' && (
                conversations.length === 0
                  ? <div className="sidebar-empty">No history yet.<br/>Start a conversation to get going.</div>
                  : conversations.map(c => (
                    <button
                      key={c.id}
                      className={`history-item${c.id === activeConvID ? ' active' : ''}`}
                      onClick={() => loadConversation(c)}
                    >
                      <span className="history-title">{c.title}</span>
                      <span className="history-meta">{c.model} · {c.updatedAt}</span>
                    </button>
                  ))
              )}

              {!loadingHistory && sidebarTab === 'memories' && (
                <>
                  <div className="memory-input-row">
                    <input
                      className="memory-input"
                      placeholder="Add a memory…"
                      value={newMemory}
                      onChange={e => setNewMemory(e.target.value)}
                      onKeyDown={e => e.key === 'Enter' && saveMemory()}
                    />
                    <button className="memory-save-btn" onClick={saveMemory} disabled={!newMemory.trim()}>+</button>
                  </div>
                  {memories.length === 0
                    ? <div className="sidebar-empty">No memories yet.<br/>Use <code>/remember</code> in the CLI<br/>or add one above.</div>
                    : memories.map(m => (
                      <div key={m.id} className="memory-item">
                        <span className="memory-content">{m.content}</span>
                        <div className="memory-footer">
                          <span className="memory-date">{m.createdAt}</span>
                          <button className="memory-delete" onClick={() => removeMemory(m.id)} title="Forget">✕</button>
                        </div>
                      </div>
                    ))
                  }
                </>
              )}
            </div>
          </div>
        </>
      )}

      {/* Ollama offline banner */}
      {ollamaOffline && (
        <div className="banner banner-warn">
          Ollama not running — start Ollama to use Billy.
        </div>
      )}

      {/* Message list */}
      <div className="messages">
        {messages.length === 0 && (
          <div className="empty-state">
            <img src={billyIcon} className="empty-icon" alt="" />
            <p>Ask Billy anything.</p>
            <p className="empty-sub">Code, explain, fix — all local, all private.</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`msg msg-${msg.role}`}>
            <div className="msg-label">{msg.role === 'user' ? 'You' : 'Billy'}</div>
            <div className="msg-content">
              {msg.content || (msg.role === 'assistant' && streaming ? <span className="cursor" /> : null)}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Model picker dropdown */}
      {showModelPicker && models.length > 0 && (
        <div className="model-picker">
          {models.map(m => (
            <button
              key={m}
              className={`model-option${m === activeModel ? ' active' : ''}`}
              onClick={() => { setActiveModel(m); setShowModelPicker(false); }}
            >
              {m}
            </button>
          ))}
        </div>
      )}

      {/* Input bar */}
      <div className="input-bar">
        <button
          className="model-btn"
          title="Switch model"
          onClick={() => setShowModelPicker(v => !v)}
        >
          {activeModel.split(':')[0]}
        </button>
        <textarea
          className="input-area"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={streaming ? 'Thinking…' : `Message Billy  (${platform === 'darwin' ? '⌘↩' : 'Ctrl+↩'} to send)`}
          disabled={streaming}
          rows={1}
        />
        <button
          className={`send-btn${streaming ? ' sending' : ''}`}
          onClick={sendMessage}
          disabled={streaming || !input.trim()}
          title="Send"
        >
          {streaming ? '…' : '↑'}
        </button>
      </div>

      {/* Status bar */}
      <div className="statusbar">
        <span className={`dot ${status?.ollamaReady ? 'dot-green' : 'dot-red'}`} />
        <span className="statusbar-model">{activeModel}</span>
        {status?.billyServing && <span className="statusbar-serving">● billy serve</span>}
        {status?.version && <span className="statusbar-ver">v{status.version}</span>}
      </div>
    </div>
  );
}

export default App;
