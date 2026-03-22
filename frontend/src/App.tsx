import { useState, useEffect, useRef } from 'react';
import billyIcon from './assets/images/billy-icon.png';
import './App.css';
import { GetStatus, SendMessage, ListModels, PopOut, OpenInstallPage, GetPlatform } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { main } from '../wailsjs/go/models';

type Theme = 'system' | 'dark' | 'light';

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
  const [showModelPicker, setShowModelPicker] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [theme, setTheme] = useState<Theme>(() => {
    return (localStorage.getItem('billy-theme') as Theme) || 'system';
  });
  const [platform, setPlatform] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);
  const streamBuffer = useRef('');

  // Apply theme on mount and whenever it changes
  useEffect(() => { applyTheme(theme); }, [theme]);

  function changeTheme(t: Theme) {
    setTheme(t);
    applyTheme(t);
    setShowSettings(false);
  }

  useEffect(() => {
    // Load status and models on mount
    GetStatus().then(s => {
      setStatus(s);
      if (s.activeModel) setActiveModel(s.activeModel);
    });
    ListModels().then(m => { if (m) setModels(m); });
    GetPlatform().then(setPlatform);

    // Subscribe to streaming events
    EventsOn('chat:token', (token: string) => {
      streamBuffer.current += token;
      setMessages(prev => {
        const msgs = [...prev];
        if (msgs.length > 0 && msgs[msgs.length - 1].role === 'assistant') {
          msgs[msgs.length - 1] = { ...msgs[msgs.length - 1], content: streamBuffer.current };
        }
        return msgs;
      });
    });

    EventsOn('chat:done', () => {
      streamBuffer.current = '';
      setStreaming(false);
    });

    EventsOn('chat:error', (err: string) => {
      streamBuffer.current = '';
      setStreaming(false);
      setMessages(prev => [...prev, { role: 'assistant', content: `⚠️ ${err}` }]);
    });

    // Refresh status every 15s
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

    SendMessage(new main.ChatRequest({ messages: newMessages, model: activeModel }));
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  function clearChat() {
    setMessages([]);
    streamBuffer.current = '';
  }

  const notReady = status && !status.ollamaReady;
  const notInstalled = status && !status.billyInstalled;

  return (
    <div className="app">
      {/* Title bar (macOS inset traffic lights area) */}
      <div className="titlebar" style={{ '--wails-draggable': 'drag' } as React.CSSProperties}>
        <div className="titlebar-left">
          <img src={billyIcon} className="titlebar-icon" alt="Billy" />
          <span className="titlebar-name">Billy</span>
          {status && (
            <span className={`tier-badge tier-${status.tier}`}>{status.tier}</span>
          )}
        </div>
        <div className="titlebar-right">
          <button className="icon-btn" title="Settings" onClick={() => setShowSettings(v => !v)}>⚙</button>
          <button className="icon-btn" title="Pop out" onClick={() => PopOut()}>⤢</button>
          <button className="icon-btn" title="Clear chat" onClick={clearChat}>⊘</button>
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

      {/* Not installed banner */}
      {notInstalled && (
        <div className="banner banner-warn">
          <span>Billy CLI not found.</span>
          <button className="banner-btn" onClick={() => OpenInstallPage()}>Install →</button>
        </div>
      )}

      {/* Ollama offline banner */}
      {notReady && !notInstalled && (
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
        {status?.billyServing && <span className="statusbar-serving">● connected</span>}
        {status?.version && <span className="statusbar-ver">v{status.version}</span>}
      </div>
    </div>
  );
}

export default App;
