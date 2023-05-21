package metadata // import "google.golang.org/grpc/metadata"

import (
	"context"
	"fmt"
	"strings"
)

func DecodeKeyValue(k, v string) (string, string, error) {
	return k, v, nil
}

// MD 类型的定义(一个map，key 是 string 类型，value 是 []string 类型)
type MD map[string][]string

// New 将 map 初始化为 MD ;
// key 的类型：数字: 0-9 ，大写字母: A-Z (标准化为小写) ，小写字母: a-z ，特殊字符:-_。
func New(m map[string]string) MD {
	md := make(MD, len(m))
	for k, val := range m {
		key := strings.ToLower(k)
		md[key] = append(md[key], val)
	}
	return md
}

// Pairs 将字符串转换为 MD ;
func Pairs(kv ...string) MD {
	// 参数个数必须得为偶数(key-value)
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kv)))
	}
	md := make(MD, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key := strings.ToLower(kv[i])
		md[key] = append(md[key], kv[i+1])
	}
	return md
}

// Join 将所有 MD 合并：相同的 key 追加；
func Join(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out[k] = append(out[k], v...)
		}
	}
	return out
}

// MD 的方法集

// Len  返回MD的长度
func (md MD) Len() int {
	return len(md)
}

// Copy 复制一个 MD 返回
func (md MD) Copy() MD {
	out := make(MD, len(md))
	for k, v := range md {
		out[k] = copyOf(v)
	}
	return out
}

// Get 通过 key 从 MD 中获取值
func (md MD) Get(k string) []string {
	k = strings.ToLower(k)
	return md[k]
}

// Set 设置/更新 key 对应的 value
func (md MD) Set(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	k = strings.ToLower(k)
	md[k] = vals
}

// Append 追加 key 对应的 value 值
func (md MD) Append(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	k = strings.ToLower(k)
	md[k] = append(md[k], vals...)
}

// Delete 删除 key-value
func (md MD) Delete(k string) {
	k = strings.ToLower(k)
	delete(md, k)
}

type mdIncomingKey struct{}
type mdOutgoingKey struct{}

// NewIncomingContext 创建一个带有传入 md 的  Context。
func NewIncomingContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdIncomingKey{}, md)
}

// NewOutgoingContext 创建一个带有传出 md 的 Context，
func NewOutgoingContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdOutgoingKey{}, rawMD{md: md})
}

// AppendToOutgoingContext 返回一个合并了 kv 的新的 Context
// 只有  rawMD 的 added 变化
func AppendToOutgoingContext(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: AppendToOutgoingContext got an odd number of input pairs for metadata: %d", len(kv)))
	}

	md, _ := ctx.Value(mdOutgoingKey{}).(rawMD)

	added := make([][]string, len(md.added)+1)
	copy(added, md.added)
	// 将所有的 string 转为小写
	kvCopy := make([]string, 0, len(kv))
	for i := 0; i < len(kv); i += 2 {
		kvCopy = append(kvCopy, strings.ToLower(kv[i]), kv[i+1])
	}

	added[len(added)-1] = kvCopy
	return context.WithValue(ctx, mdOutgoingKey{}, rawMD{md: md.md, added: added})
}

// FromIncomingContext 返回 ctx 中的传入元数据(如果存在)， MD 中的所有键都是小写；
func FromIncomingContext(ctx context.Context) (MD, bool) {
	md, ok := ctx.Value(mdIncomingKey{}).(MD)
	if !ok {
		return nil, false
	}
	out := make(MD, len(md))
	for k, v := range md {
		// We need to manually convert all keys to lower case, because MD is a
		// map, and there's no guarantee that the MD attached to the context is
		// created using our helper functions.
		key := strings.ToLower(k)
		out[key] = copyOf(v)
	}
	return out, true
}

// ValueFromIncomingContext ：通过 key 从 ctx 中的 MD 获取相应的值（如果值存在）；
func ValueFromIncomingContext(ctx context.Context, key string) []string {
	md, ok := ctx.Value(mdIncomingKey{}).(MD)
	if !ok {
		return nil
	}
	// 直接通过 key 从 MD 找值
	if v, ok := md[key]; ok {
		return copyOf(v)
	}
	// 将 key 转为小写，从 MD 找值
	for k, v := range md {
		// We need to manually convert all keys to lower case, because MD is a
		// map, and there's no guarantee that the MD attached to the context is
		// created using our helper functions.
		if strings.ToLower(k) == key {
			return copyOf(v)
		}
	}
	return nil
}

// copyOf 拷贝一个新的 slice
func copyOf(v []string) []string {
	vals := make([]string, len(v))
	copy(vals, v)
	return vals
}

// FromOutgoingContextRaw 返回 Raw 的 MD []string(对 MD 的 key 没有进行处理)；如果不存在返回 false
func FromOutgoingContextRaw(ctx context.Context) (MD, [][]string, bool) {
	raw, ok := ctx.Value(mdOutgoingKey{}).(rawMD)
	if !ok {
		return nil, nil, false
	}

	return raw.md, raw.added, true
}

// FromOutgoingContext 返回 ctx 中的传出元数据(如果存在)，返回的 MD 中的所有键都是小写。
func FromOutgoingContext(ctx context.Context) (MD, bool) {
	raw, ok := ctx.Value(mdOutgoingKey{}).(rawMD)
	if !ok {
		return nil, false
	}

	mdSize := len(raw.md)
	for i := range raw.added {
		mdSize += len(raw.added[i]) / 2
	}

	out := make(MD, mdSize)
	for k, v := range raw.md {
		// We need to manually convert all keys to lower case, because MD is a
		// map, and there's no guarantee that the MD attached to the context is
		// created using our helper functions.
		key := strings.ToLower(k)
		out[key] = copyOf(v)
	}

	for _, added := range raw.added {
		if len(added)%2 == 1 {
			panic(fmt.Sprintf("metadata: FromOutgoingContext got an odd number of input pairs for metadata: %d", len(added)))
		}

		for i := 0; i < len(added); i += 2 {
			key := strings.ToLower(added[i])
			out[key] = append(out[key], added[i+1])
		}
	}
	return out, ok
}

type rawMD struct {
	md    MD
	added [][]string
}
