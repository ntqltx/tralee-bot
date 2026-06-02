const TELEGRAM_API = "https://api.telegram.org";

export default {
  async fetch(request, env) {
    if (request.method !== "POST") {
      return new Response("ok", { status: 200 });
    }

    let update;
    try {
      update = await request.json();
    } catch {
      return new Response("bad json", { status: 400 });
    }

    const msg = update?.message;
    const chatId = msg?.chat?.id;
    const text = (msg?.text || "").trim();

    if (!chatId || !text) {
      return new Response("ok", { status: 200 });
    }

    const key = `sub:${chatId}`;

    if (text === "/start") {
      const existing = await env.KV.get(key);
      if (existing) {
        await reply(env, chatId, "ℹ️ Already subscribed to Tralee apartment alerts. Send /stop to unsubscribe.");
      } 
      else {
        await env.KV.put(key, String(chatId));
        await reply(env, chatId, "✅ You've subscribed to Tralee apartment alerts! Send /stop to unsubscribe.");
      }
    } 
    else if (text === "/stop") {
      const existing = await env.KV.get(key);
      if (!existing) {
        await reply(env, chatId, "ℹ️ You're not subscribed. Send /start to subscribe.");
      } 
      else {
        await env.KV.delete(key);
        await reply(env, chatId, "❌ Unsubscribed. Send /start to resubscribe.");
      }
    }

    return new Response("ok", { status: 200 });
  },
};

async function reply(env, chatId, text) {
  const url = `${TELEGRAM_API}/bot${env.BOT_TOKEN}/sendMessage`;
  await fetch(url, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ chat_id: chatId, text }),
  });
}
