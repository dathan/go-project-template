import { useState, useRef, useEffect } from "react";
import { streamURL, getToken } from "../api/client";
import Layout from "../components/Layout";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
}

export default function AgentChat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const appendAssistant = (chunk: string) => {
    setMessages((prev) => {
      const last = prev.at(-1);
      if (last?.role === "assistant") {
        return [...prev.slice(0, -1), { ...last, content: last.content + chunk }];
      }
      return [...prev, { id: crypto.randomUUID(), role: "assistant", content: chunk }];
    });
  };

  const sendMessage = async () => {
    const prompt = input.trim();
    if (!prompt || streaming) return;

    setInput("");
    setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: "user", content: prompt }]);
    setStreaming(true);

    const token = getToken();
    const url = streamURL(prompt) + (token ? `&token=${token}` : "");
    const es = new EventSource(url);

    es.onmessage = (e) => {
      try {
        const text = JSON.parse(e.data) as string;
        appendAssistant(text);
      } catch {
        appendAssistant(e.data);
      }
    };

    es.addEventListener("done", () => {
      es.close();
      setStreaming(false);
    });

    es.onerror = () => {
      es.close();
      setStreaming(false);
    };
  };

  return (
    <Layout>
      <div className="flex flex-col h-[calc(100vh-10rem)] max-w-3xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">Agent Chat</h1>

        <div className="flex-1 overflow-y-auto bg-white rounded-xl border border-gray-200 p-4 space-y-4">
          {messages.length === 0 && (
            <p className="text-center text-gray-400 text-sm mt-8">
              Ask the agent anything. It has access to shell, filesystem, browser, and HTTP tools.
            </p>
          )}
          {messages.map((m) => (
            <div key={m.id} className={`flex ${m.role === "user" ? "justify-end" : "justify-start"}`}>
              <div className={`max-w-[80%] rounded-2xl px-4 py-2.5 text-sm whitespace-pre-wrap ${
                m.role === "user"
                  ? "bg-indigo-600 text-white rounded-br-none"
                  : "bg-gray-100 text-gray-900 rounded-bl-none"
              }`}>
                {m.content}
              </div>
            </div>
          ))}
          {streaming && (
            <div className="flex justify-start">
              <div className="bg-gray-100 rounded-2xl rounded-bl-none px-4 py-2.5">
                <span className="inline-flex gap-1">
                  {[0, 1, 2].map((i) => (
                    <span key={i} className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: `${i * 150}ms` }} />
                  ))}
                </span>
              </div>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        <form
          onSubmit={(e) => { e.preventDefault(); sendMessage(); }}
          className="mt-3 flex gap-2"
        >
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type a message…"
            className="flex-1 border border-gray-300 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <button
            type="submit"
            disabled={streaming || !input.trim()}
            className="bg-indigo-600 text-white px-5 py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
          >
            Send
          </button>
        </form>
      </div>
    </Layout>
  );
}
